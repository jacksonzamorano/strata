package componentipc

import "encoding/json"

type ComponentMessage struct {
	Id      string          `json:"id"`
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ComponentResultPayload struct {
	Success  bool
	Response json.RawMessage
	Error    string
}

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

type ComponentMessageLog struct {
	Message string `json:"message"`
}

type ComponentMessageGetKeychainRequest struct {
	Key string `json:"key"`
}

type ComponentMessageGetKeychainResponse struct {
	Value string `json:"value"`
}

type ComponentMessageSetKeychainRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ComponentMessageSendTrigger struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}
