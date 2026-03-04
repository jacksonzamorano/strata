package hostio

import "encoding/json"

type HostMessagePayload = json.RawMessage

type HostMessage struct {
	Id      string          `json:"id"`
	Type    HostMessageType `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func NewHostMessage(id string, typ HostMessageType, payload any) (HostMessage, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return HostMessage{}, err
	}
	return HostMessage{
		Id:      id,
		Type:    typ,
		Payload: encoded,
	}, nil
}

func DecodeHostMessagePayload[T any](msg HostMessage) (T, error) {
	var payload T
	err := json.Unmarshal(msg.Payload, &payload)
	return payload, err
}
