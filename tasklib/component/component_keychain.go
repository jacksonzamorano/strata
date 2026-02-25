package component

import "github.com/jacksonzamorano/tasks/tasklib/internal/componentipc"

type ComponentKeychain struct {
	io *componentipc.IO
}

func newComponentKeychain(io *componentipc.IO) *ComponentKeychain {
	return &ComponentKeychain{io: io}
}

func (c *ComponentKeychain) Get(k string) string {
	thread := c.io.NewThread()
	payload, _ := componentipc.SendAndReceive[componentipc.ComponentMessageGetKeychainResponse](
		thread,
		componentipc.MessageTypeGetKeychainRequest,
		componentipc.ComponentMessageGetKeychainRequest{Key: k},
		componentipc.MessageTypeGetKeychainResponse,
	)
	return payload.Value
}

func (c *ComponentKeychain) Set(k, v string) {
	c.io.NewThread().Send(componentipc.MessageTypeStoreKeychainRequest, componentipc.ComponentMessageSetKeychainRequest{
		Key:   k,
		Value: v,
	})
}
