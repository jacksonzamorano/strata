package component

import "encoding/json"

type ComponentMessageHello struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type ComponentMessageReady struct {
	Error string `json:"error"`
}
type ComponentMessageGetValueRequest struct {
	Key string `json:"key"`
}
type ComponentMessageGetValueResponse struct {
	Value string `json:"value"`
}
type ComponentMessageSetValueRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type ComponentMessageExecute struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
