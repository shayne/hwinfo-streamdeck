package streamdeck

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// EventDelegate receives callbacks for Stream Deck SDK events
type EventDelegate interface {
	OnConnected(*websocket.Conn)
	OnWillAppear(*EvWillAppear)
	OnTitleParametersDidChange(*EvTitleParametersDidChange)
	OnPropertyInspectorConnected(*EvSendToPlugin)
	OnSendToPlugin(*EvSendToPlugin)
	OnApplicationDidLaunch(*EvApplication)
	OnApplicationDidTerminate(*EvApplication)
}

// StreamDeck SDK APIs
type StreamDeck struct {
	Port          string
	PluginUUID    string
	RegisterEvent string
	Info          string
	delegate      EventDelegate
	conn          *websocket.Conn
	done          chan struct{}
}

// NewStreamDeck prepares StreamDeck struct
func NewStreamDeck(port, pluginUUID, registerEvent, info string) *StreamDeck {
	return &StreamDeck{
		Port:          port,
		PluginUUID:    pluginUUID,
		RegisterEvent: registerEvent,
		Info:          info,
		done:          make(chan struct{}),
	}
}

// SetDelegate sets the delegate for receiving Stream Deck SDK event callbacks
func (sd *StreamDeck) SetDelegate(ed EventDelegate) {
	sd.delegate = ed
}

func (sd *StreamDeck) register() error {
	reg := evRegister{Event: sd.RegisterEvent, UUID: sd.PluginUUID}
	data, err := json.Marshal(reg)
	log.Println(string(data))
	if err != nil {
		return err
	}
	err = sd.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}
	return nil
}

// Connect establishes WebSocket connection to StreamDeck software
func (sd *StreamDeck) Connect() error {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("127.0.0.1:%s", sd.Port)}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	sd.conn = c

	err = sd.register()
	if err != nil {
		return fmt.Errorf("failed register: %v", err)
	}

	if sd.delegate != nil {
		sd.delegate.OnConnected(sd.conn)
	}

	return nil
}

// Close closes the websocket connection, defer after Connect
func (sd *StreamDeck) Close() {
	sd.conn.Close()
}

func (sd *StreamDeck) onPropertyInspectorMessage(value string, ev *EvSendToPlugin) error {
	switch value {
	case "propertyInspectorConnected":
		if sd.delegate != nil {
			sd.delegate.OnPropertyInspectorConnected(ev)
		}
	default:
		log.Printf("Unknown property_inspector value: %s\n", value)
	}
	return nil
}

func (sd *StreamDeck) onSendToPlugin(ev *EvSendToPlugin) error {
	payload := make(map[string]*json.RawMessage)
	err := json.Unmarshal(*ev.Payload, &payload)
	if err != nil {
		return fmt.Errorf("onSendToPlugin payload unmarshal: %v", err)
	}
	if raw, ok := payload["property_inspector"]; ok {
		var value string
		err := json.Unmarshal(*raw, &value)
		if err != nil {
			return fmt.Errorf("onSendToPlugin unmarshal property_inspector value: %v", err)
		}
		err = sd.onPropertyInspectorMessage(value, ev)
		if err != nil {
			return fmt.Errorf("onPropertyInspectorMessage: %v", err)
		}
		return nil
	}
	if sd.delegate != nil {
		sd.delegate.OnSendToPlugin(ev)
	}
	return nil
}

func (sd *StreamDeck) spawnMessageReader() {
	defer close(sd.done)
	for {
		_, message, err := sd.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)

		var objmap map[string]*json.RawMessage
		err = json.Unmarshal(message, &objmap)
		if err != nil {
			log.Fatal("message unmarshal", err)
		}
		var event string
		err = json.Unmarshal(*objmap["event"], &event)
		if err != nil {
			log.Fatal("event unmarshal", err)
		}
		switch event {
		case "willAppear":
			var ev EvWillAppear
			err := json.Unmarshal(message, &ev)
			if err != nil {
				log.Fatal("willAppear unmarshal", err)
			}
			if sd.delegate != nil {
				sd.delegate.OnWillAppear(&ev)
			}
		case "titleParametersDidChange":
			var ev EvTitleParametersDidChange
			err := json.Unmarshal(message, &ev)
			if err != nil {
				log.Fatal("titleParametersDidChange unmarshal", err)
			}
			if sd.delegate != nil {
				sd.delegate.OnTitleParametersDidChange(&ev)
			}
		case "sendToPlugin":
			var ev EvSendToPlugin
			err := json.Unmarshal(message, &ev)
			if err != nil {
				log.Fatal("onSendToPlugin event unmarshal", err)
			}
			err = sd.onSendToPlugin(&ev)
			if err != nil {
				log.Fatal("onSendToPlugin", err)
			}
		case "applicationDidLaunch":
			var ev EvApplication
			err := json.Unmarshal(message, &ev)
			if err != nil {
				log.Fatal("applicationDidLaunch unmarshal", err)
			}
			if sd.delegate != nil {
				sd.delegate.OnApplicationDidLaunch(&ev)
			}
		case "applicationDidTerminate":
			var ev EvApplication
			err := json.Unmarshal(message, &ev)
			if err != nil {
				log.Fatal("applicationDidTerminate unmarshal", err)
			}
			if sd.delegate != nil {
				sd.delegate.OnApplicationDidTerminate(&ev)
			}
		default:
			log.Printf("Unknown event: %s\n", event)
		}
	}
}

// ListenAndWait processes messages and waits until closed
func (sd *StreamDeck) ListenAndWait() {
	go sd.spawnMessageReader()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-sd.done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := sd.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-sd.done:
			case <-time.After(time.Second):
			}
		}
	}
}

// SendToPropertyInspector sends a payload to the Property Inspector
func (sd *StreamDeck) SendToPropertyInspector(action, context string, payload interface{}) error {
	event := evSendToPropertyInspector{Action: action, Event: "sendToPropertyInspector",
		Context: context, Payload: payload}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("sendToPropertyInspector: %v", err)
	}
	err = sd.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("setTitle write: %v", err)
	}
	return nil
}

// SetTitle dynamically changes the title displayed by an instance of an action
func (sd *StreamDeck) SetTitle(context, title string) error {
	event := evSetTitle{Event: "setTitle", Context: context, Payload: evSetTitlePayload{
		Title:  title,
		Target: 0,
	}}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("setTitle: %v", err)
	}
	err = sd.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("setTitle write: %v", err)
	}
	return nil
}

// SetSettings saves persistent data for the action's instance
func (sd *StreamDeck) SetSettings(context string, payload interface{}) error {
	event := evSetSettings{Event: "setSettings", Context: context, Payload: payload}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("setSettings: %v", err)
	}
	err = sd.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("setSettings write: %v", err)
	}
	return nil
}

// SetImage dynamically changes the image displayed by an instance of an action
func (sd *StreamDeck) SetImage(context string, bts []byte) error {
	b64 := base64.StdEncoding.EncodeToString(bts)
	event := evSetImage{Event: "setImage", Context: context, Payload: evSetImagePayload{
		Image:  fmt.Sprintf("data:image/png;base64, %s", b64),
		Target: 0,
	}}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("setImage: %v", err)
	}
	err = sd.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("setImage write: %v", err)
	}
	return nil
}
