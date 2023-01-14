package hwsensorsservice

import (
	"context"

	"google.golang.org/grpc"

	"github.com/hashicorp/go-plugin"
	"github.com/shayne/hwinfo-streamdeck/pkg/service/proto"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"hwinfoplugin": &HardwareServicePlugin{},
}

// HardwareService is the interface that we're exposing as a plugin.
type HardwareService interface {
	PollTime() (uint64, error)
	Sensors() ([]Sensor, error)
	ReadingsForSensorID(id string) ([]Reading, error)
}

// HardwareServicePlugin is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type HardwareServicePlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl HardwareService
}

// GRPCServer constructor
func (p *HardwareServicePlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterHWServiceServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient constructor
func (p *HardwareServicePlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{Client: proto.NewHWServiceClient(c)}, nil
}

// Sensor is the common hardware interface for a sensor
type Sensor interface {
	ID() string
	Name() string
}

// ReadingType enum of value/unit type for reading
type ReadingType int

const (
	// ReadingTypeNone no type
	ReadingTypeNone ReadingType = iota
	// ReadingTypeTemp temperature in celsius
	ReadingTypeTemp
	// ReadingTypeVolt voltage
	ReadingTypeVolt
	// ReadingTypeFan RPM
	ReadingTypeFan
	// ReadingTypeCurrent amps
	ReadingTypeCurrent
	// ReadingTypePower watts
	ReadingTypePower
	// ReadingTypeClock Mhz
	ReadingTypeClock
	// ReadingTypeUsage e.g. MBs
	ReadingTypeUsage
	// ReadingTypeOther other
	ReadingTypeOther
)

func (t ReadingType) String() string {
	return [...]string{"None", "Temp", "Volt", "Fan", "Current", "Power", "Clock", "Usage", "Other"}[t]
}

// Reading is the common hardware interface for a sensor's reading
type Reading interface {
	ID() int32
	TypeI() int32
	Type() string
	Label() string
	Unit() string
	Value() float64
	ValueMin() float64
	ValueMax() float64
	ValueAvg() float64
}

type sensor struct {
	*proto.Sensor
}

func (s sensor) ID() string {
	return s.Sensor.GetID()
}

func (s sensor) Name() string {
	return s.Sensor.GetName()
}

type reading struct {
	*proto.Reading
}

func (r reading) ID() int32 {
	return r.Reading.GetID()
}

func (r reading) Label() string {
	return r.Reading.GetLabel()
}

func (r reading) Type() string {
	return r.Reading.GetType()
}

func (r reading) TypeI() int32 {
	return r.Reading.GetTypeI()
}

func (r reading) Unit() string {
	return r.Reading.GetUnit()
}

func (r reading) Value() float64 {
	return r.Reading.GetValue()
}

func (r reading) ValueMin() float64 {
	return r.Reading.GetValueMin()
}

func (r reading) ValueMax() float64 {
	return r.Reading.GetValueMax()
}

func (r reading) ValueAvg() float64 {
	return r.Reading.GetValueAvg()
}
