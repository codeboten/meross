package api

type mqttHeader struct {
	From           string `json:"from"`
	MessageID      string `json:"messageId"`
	Method         string `json:"method"`
	Namespace      string `json:"namespace"`
	PayloadVersion int    `json:"payloadVersion"`
	Signature      string `json:"sign"`
	Timestamp      int    `json:"timestamp"`
}

type mqttMessage struct {
	Header  mqttHeader  `json:"header"`
	Payload mqttPayload `json:"payload"`
}
type mqttPayload struct {
	Toggle map[string]int `json:"togglex"`
}
