package hwinfostreamdeckplugin

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/shayne/go-winpeg"
	"github.com/shayne/hwinfo-streamdeck/pkg/graph"
	hwsensorsservice "github.com/shayne/hwinfo-streamdeck/pkg/service"
	"github.com/shayne/hwinfo-streamdeck/pkg/streamdeck"
)

// Plugin handles information between HWiNFO and Stream Deck
type Plugin struct {
	c      *plugin.Client
	peg    winpeg.ProcessExitGroup
	hw     hwsensorsservice.HardwareService
	sd     *streamdeck.StreamDeck
	am     *actionManager
	graphs map[string]*graph.Graph

	appLaunched bool
}

func (p *Plugin) startClient() error {
	cmd := exec.Command("./hwinfo-plugin.exe")

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  hwsensorsservice.Handshake,
		Plugins:          hwsensorsservice.PluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		AutoMTLS:         true,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return err
	}

	g, err := winpeg.NewProcessExitGroup()
	if err != nil {
		return err
	}

	if err := g.AddProcess(cmd.Process); err != nil {
		return err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("hwinfoplugin")
	if err != nil {
		return err
	}

	p.c = client
	p.peg = g
	p.hw = raw.(hwsensorsservice.HardwareService)

	return nil
}

// NewPlugin creates an instance and initializes the plugin
func NewPlugin(port, uuid, event, info string) (*Plugin, error) {
	// We don't want to see the plugin logs.
	// log.SetOutput(ioutil.Discard)
	p := &Plugin{
		am:     newActionManager(),
		graphs: make(map[string]*graph.Graph),
	}
	p.startClient()
	p.sd = streamdeck.NewStreamDeck(port, uuid, event, info)
	return p, nil
}

// RunForever starts the plugin and waits for events, indefinitely
func (p *Plugin) RunForever() error {
	defer func() {
		p.c.Kill()
		p.peg.Dispose()
	}()

	p.sd.SetDelegate(p)
	p.am.Run(p.updateTiles)

	go func() {
		for {
			if p.c.Exited() {
				p.startClient()
			}
			time.Sleep(1 * time.Second)
		}
	}()

	err := p.sd.Connect()
	if err != nil {
		return fmt.Errorf("StreamDeck Connect: %v", err)
	}
	defer p.sd.Close()
	p.sd.ListenAndWait()
	return nil
}

func (p *Plugin) getReading(suid string, rid int32) (hwsensorsservice.Reading, error) {
	rbs, err := p.hw.ReadingsForSensorID(suid)
	if err != nil {
		return nil, fmt.Errorf("getReading ReadingsBySensor failed: %v", err)
	}
	for _, r := range rbs {
		if r.ID() == rid {
			return r, nil
		}
	}
	return nil, fmt.Errorf("ReadingID does not exist: %s", suid)
}

func (p *Plugin) applyDefaultFormat(v float64, t hwsensorsservice.ReadingType, u string) string {
	switch t {
	case hwsensorsservice.ReadingTypeNone:
		return fmt.Sprintf("%0.f %s", v, u)
	case hwsensorsservice.ReadingTypeTemp:
		return fmt.Sprintf("%.0f %s", v, u)
	case hwsensorsservice.ReadingTypeVolt:
		return fmt.Sprintf("%.0f %s", v, u)
	case hwsensorsservice.ReadingTypeFan:
		return fmt.Sprintf("%.0f %s", v, u)
	case hwsensorsservice.ReadingTypeCurrent:
		return fmt.Sprintf("%.0f %s", v, u)
	case hwsensorsservice.ReadingTypePower:
		return fmt.Sprintf("%0.f %s", v, u)
	case hwsensorsservice.ReadingTypeClock:
		return fmt.Sprintf("%.0f %s", v, u)
	case hwsensorsservice.ReadingTypeUsage:
		return fmt.Sprintf("%.0f%s", v, u)
	case hwsensorsservice.ReadingTypeOther:
		return fmt.Sprintf("%.0f %s", v, u)
	}
	return "Bad Format"
}

func (p *Plugin) updateTiles(data *actionData) {
	if data.action != "com.exension.hwinfo.reading" {
		log.Printf("Unknown action updateTiles: %s\n", data.action)
		return
	}

	g, ok := p.graphs[data.context]
	if !ok {
		log.Printf("Graph not found for context: %s\n", data.context)
		return
	}

	if !p.appLaunched {
		if !data.settings.InErrorState {
			payload := evStatus{Error: true, Message: "HWiNFO Unavailable"}
			err := p.sd.SendToPropertyInspector("com.exension.hwinfo.reading", data.context, payload)
			if err != nil {
				log.Println("updateTiles SendToPropertyInspector", err)
			}
			data.settings.InErrorState = true
			p.sd.SetSettings(data.context, &data.settings)
		}
		bts, err := ioutil.ReadFile("./launch-hwinfo.png")
		if err != nil {
			log.Printf("Failed to read launch-hwinfo.png: %v\n", err)
			return
		}
		err = p.sd.SetImage(data.context, bts)
		if err != nil {
			log.Printf("Failed to setImage: %v\n", err)
			return
		}
		return
	}

	// show ui on property inspector if in error state
	if data.settings.InErrorState {
		payload := evStatus{Error: false, Message: "show_ui"}
		err := p.sd.SendToPropertyInspector("com.exension.hwinfo.reading", data.context, payload)
		if err != nil {
			log.Println("updateTiles SendToPropertyInspector", err)
		}
		data.settings.InErrorState = false
		p.sd.SetSettings(data.context, &data.settings)
	}

	s := data.settings
	r, err := p.getReading(s.SensorUID, s.ReadingID)
	if err != nil {
		log.Printf("getReading failed: %v\n", err)
		return
	}

	v := r.Value()
	if s.Divisor != "" {
		fdiv := 1.
		fdiv, err := strconv.ParseFloat(s.Divisor, 64)
		if err != nil {
			log.Printf("Failed to parse float: %s\n", s.Divisor)
			return
		}
		v = r.Value() / fdiv
	}
	g.Update(v)
	var text string
	if f := s.Format; f != "" {
		text = fmt.Sprintf(f, v)
	} else {
		text = p.applyDefaultFormat(v, hwsensorsservice.ReadingType(r.TypeI()), r.Unit())
	}
	g.SetLabelText(1, text)

	b, err := g.EncodePNG()
	if err != nil {
		log.Printf("Failed to encode graph: %v\n", err)
		return
	}

	err = p.sd.SetImage(data.context, b)
	if err != nil {
		log.Printf("Failed to setImage: %v\n", err)
		return
	}
}
