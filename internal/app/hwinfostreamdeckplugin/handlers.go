package hwinfostreamdeckplugin

import (
	"fmt"
	"image/color"
	"strconv"

	hwsensorsservice "github.com/shayne/go-hwinfo-streamdeck-plugin/pkg/service"
	"github.com/shayne/go-hwinfo-streamdeck-plugin/pkg/streamdeck"
)

func (p *Plugin) handleSensorSelect(event *streamdeck.EvSendToPlugin, sdpi *evSdpiCollection) error {
	sensorid := sdpi.Value
	readings, err := p.hw.ReadingsForSensorID(sensorid)
	if err != nil {
		return fmt.Errorf("handleSensorSelect ReadingsBySensor failed: %v", err)
	}
	evreadings := []*evSendReadingsPayloadReading{}
	for _, r := range readings {
		evreadings = append(evreadings, &evSendReadingsPayloadReading{ID: r.ID(), Label: r.Label(), Prefix: r.Unit()})
	}
	settings, err := p.am.getSettings(event.Context)
	if err != nil {
		return fmt.Errorf("handleReadingSelect getSettings: %v", err)
	}
	// only update settings if SensorUID is changing
	// this covers case where PI sends event when tile
	// selected in SD UI
	if settings.SensorUID != sensorid {
		settings.SensorUID = sensorid
		settings.ReadingID = 0
		settings.IsValid = false
	}
	payload := evSendReadingsPayload{Readings: evreadings, Settings: &settings}
	err = p.sd.SendToPropertyInspector(event.Action, event.Context, payload)
	if err != nil {
		return fmt.Errorf("sensorsSelect SendToPropertyInspector: %v", err)
	}
	err = p.sd.SetSettings(event.Context, &settings)
	if err != nil {
		return fmt.Errorf("handleSensorSelect SetSettings: %v", err)
	}
	p.am.SetAction(event.Action, event.Context, &settings)
	return nil
}

func getDefaultMinMaxForReading(r hwsensorsservice.Reading) (int, int) {
	switch r.Unit() {
	case "%":
		return 0, 100
	case "Yes/No":
		return 0, 1
	}
	min := r.ValueMin()
	max := r.ValueMax()
	min -= min * .2
	if min <= 0 {
		min = 0.
	}
	max += max * .2
	return int(min), int(max)
}

func (p *Plugin) handleReadingSelect(event *streamdeck.EvSendToPlugin, sdpi *evSdpiCollection) error {
	rid64, err := strconv.ParseInt(sdpi.Value, 10, 32)
	if err != nil {
		return fmt.Errorf("handleReadingSelect Atoi failed: %s, %v", sdpi.Value, err)
	}
	rid := int32(rid64)
	settings, err := p.am.getSettings(event.Context)
	if err != nil {
		return fmt.Errorf("handleReadingSelect getSettings: %v", err)
	}

	// no action if reading didn't change
	if settings.ReadingID == rid {
		return nil
	}

	settings.ReadingID = rid

	// set default min/max
	r, err := p.getReading(settings.SensorUID, settings.ReadingID)
	if err != nil {
		return fmt.Errorf("handleReadingSelect getReading: %v", err)
	}

	g, ok := p.graphs[event.Context]
	if !ok {
		return fmt.Errorf("handleReadingSelect no graph for context: %s", event.Context)
	}
	defaultMin, defaultMax := getDefaultMinMaxForReading(r)
	settings.Min = defaultMin
	g.SetMin(settings.Min)
	settings.Max = defaultMax
	g.SetMax(settings.Max)
	settings.IsValid = true // set IsValid once we choose reading

	err = p.sd.SetSettings(event.Context, &settings)
	if err != nil {
		return fmt.Errorf("handleReadingSelect SetSettings: %v", err)
	}
	p.am.SetAction(event.Action, event.Context, &settings)
	return nil
}

func (p *Plugin) handleSetMin(event *streamdeck.EvSendToPlugin, sdpi *evSdpiCollection) error {
	min, err := strconv.Atoi(sdpi.Value)
	if err != nil {
		return fmt.Errorf("handleSetMin strconv: %v", err)
	}
	g, ok := p.graphs[event.Context]
	if !ok {
		return fmt.Errorf("handleSetMax no graph for context: %s", event.Context)
	}
	g.SetMin(min)
	settings, err := p.am.getSettings(event.Context)
	if err != nil {
		return fmt.Errorf("handleSetMin getSettings: %v", err)
	}
	settings.Min = min
	err = p.sd.SetSettings(event.Context, &settings)
	if err != nil {
		return fmt.Errorf("handleSetMin SetSettings: %v", err)
	}
	p.am.SetAction(event.Action, event.Context, &settings)
	return nil
}

func (p *Plugin) handleSetMax(event *streamdeck.EvSendToPlugin, sdpi *evSdpiCollection) error {
	max, err := strconv.Atoi(sdpi.Value)
	if err != nil {
		return fmt.Errorf("handleSetMax strconv: %v", err)
	}
	g, ok := p.graphs[event.Context]
	if !ok {
		return fmt.Errorf("handleSetMax no graph for context: %s", event.Context)
	}
	g.SetMax(max)
	settings, err := p.am.getSettings(event.Context)
	if err != nil {
		return fmt.Errorf("handleSetMax getSettings: %v", err)
	}
	settings.Max = max
	err = p.sd.SetSettings(event.Context, &settings)
	if err != nil {
		return fmt.Errorf("handleSetMax SetSettings: %v", err)
	}
	p.am.SetAction(event.Action, event.Context, &settings)
	return nil
}

func (p *Plugin) handleSetFormat(event *streamdeck.EvSendToPlugin, sdpi *evSdpiCollection) error {
	format := sdpi.Value
	settings, err := p.am.getSettings(event.Context)
	if err != nil {
		return fmt.Errorf("handleSetFormat getSettings: %v", err)
	}
	settings.Format = format
	err = p.sd.SetSettings(event.Context, &settings)
	if err != nil {
		return fmt.Errorf("handleSetFormat SetSettings: %v", err)
	}
	p.am.SetAction(event.Action, event.Context, &settings)
	return nil
}

func (p *Plugin) handleDivisor(event *streamdeck.EvSendToPlugin, sdpi *evSdpiCollection) error {
	divisor := sdpi.Value
	settings, err := p.am.getSettings(event.Context)
	if err != nil {
		return fmt.Errorf("handleDivisor getSettings: %v", err)
	}
	settings.Divisor = divisor
	err = p.sd.SetSettings(event.Context, &settings)
	if err != nil {
		return fmt.Errorf("handleDivisor SetSettings: %v", err)
	}
	p.am.SetAction(event.Action, event.Context, &settings)
	return nil
}

const (
	hexFormat      = "#%02x%02x%02x"
	hexShortFormat = "#%1x%1x%1x"
	hexToRGBFactor = 17
)

func hexToRGBA(hex string) *color.RGBA {
	var r, g, b uint8

	if len(hex) == 4 {
		fmt.Sscanf(hex, hexShortFormat, &r, &g, &b)
		r *= hexToRGBFactor
		g *= hexToRGBFactor
		b *= hexToRGBFactor
	} else {
		fmt.Sscanf(hex, hexFormat, &r, &g, &b)
	}

	return &color.RGBA{R: r, G: g, B: b, A: 255}
}

func (p *Plugin) handleColorChange(event *streamdeck.EvSendToPlugin, key string, sdpi *evSdpiCollection) error {
	hex := sdpi.Value
	settings, err := p.am.getSettings(event.Context)
	if err != nil {
		return fmt.Errorf("handleDivisor getSettings: %v", err)
	}
	g, ok := p.graphs[event.Context]
	if !ok {
		return fmt.Errorf("handleSetMax no graph for context: %s", event.Context)
	}
	clr := hexToRGBA(hex)
	switch key {
	case "foreground":
		settings.ForegroundColor = hex
		g.SetForegroundColor(clr)
	case "background":
		settings.BackgroundColor = hex
		g.SetBackgroundColor(clr)
	case "highlight":
		settings.HighlightColor = hex
		g.SetHighlightColor(clr)
	case "valuetext":
		settings.ValueTextColor = hex
		g.SetLabelColor(1, clr)
	}
	err = p.sd.SetSettings(event.Context, &settings)
	if err != nil {
		return fmt.Errorf("handleColorChange SetSettings: %v", err)
	}
	p.am.SetAction(event.Action, event.Context, &settings)
	return nil
}

func (p *Plugin) handleSetFontSize(event *streamdeck.EvSendToPlugin, key string, sdpi *evSdpiCollection) error {
	sv := sdpi.Value
	size, err := strconv.ParseFloat(sv, 64)
	if err != nil {
		return fmt.Errorf("failed to convert value to float: %w", err)
	}

	settings, err := p.am.getSettings(event.Context)
	if err != nil {
		return fmt.Errorf("getSettings failed: %w", err)
	}

	g, ok := p.graphs[event.Context]
	if !ok {
		return fmt.Errorf("no graph for context: %s", event.Context)
	}

	switch key {
	case "titleFontSize":
		settings.TitleFontSize = size
		g.SetLabelFontSize(0, size)
	case "valueFontSize":
		settings.ValueFontSize = size
		g.SetLabelFontSize(1, size)
	default:
		return fmt.Errorf("invalid key: %s", sdpi.Key)
	}

	err = p.sd.SetSettings(event.Context, &settings)
	if err != nil {
		return fmt.Errorf("SetSettings failed: %w", err)
	}

	p.am.SetAction(event.Action, event.Context, &settings)
	return nil
}
