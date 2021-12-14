package hwinfostreamdeckplugin

type actionSettings struct {
	SensorUID       string  `json:"sensorUid"`
	ReadingID       int32   `json:"readingId,string"`
	Title           string  `json:"title"`
	TitleFontSize   float64 `json:"titleFontSize"`
	ValueFontSize   float64 `json:"valueFontSize"`
	Min             int     `json:"min"`
	Max             int     `json:"max"`
	Format          string  `json:"format"`
	Divisor         string  `json:"divisor"`
	IsValid         bool    `json:"isValid"`
	TitleColor      string  `json:"titleColor"`
	ForegroundColor string  `json:"foregroundColor"`
	BackgroundColor string  `json:"backgroundColor"`
	HighlightColor  string  `json:"highlightColor"`
	ValueTextColor  string  `json:"valueTextColor"`
	InErrorState    bool    `json:"inErrorState"`
}

type actionData struct {
	action   string
	context  string
	settings *actionSettings
}

type evStatus struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type evSendSensorsPayloadSensor struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type evSendSensorsPayload struct {
	Sensors  []*evSendSensorsPayloadSensor `json:"sensors"`
	Settings *actionSettings               `json:"settings"`
}

type evSendReadingsPayloadReading struct {
	ID     int32  `json:"id,string"`
	Label  string `json:"label"`
	Prefix string `json:"prefix"`
}

type evSendReadingsPayload struct {
	Readings []*evSendReadingsPayloadReading `json:"readings"`
	Settings *actionSettings                 `json:"settings"`
}

type evSdpiCollection struct {
	Group     bool     `json:"group"`
	Index     int      `json:"index"`
	Key       string   `json:"key"`
	Selection []string `json:"selection"`
	Value     string   `json:"value"`
}
