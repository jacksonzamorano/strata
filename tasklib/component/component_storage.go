package component

type ComponentStorage struct {
	io *ComponentIO
}

func newComponentStorage(io *ComponentIO) *ComponentStorage {
	return &ComponentStorage{
		io,
	}
}

func (c *ComponentStorage) getValue(k string) string {
	thread := c.io.NewThread()
	payload, _ := SendAndReceive[ComponentMessageGetValueResponse](thread, ComponentMessageTypeGetValueRequest, ComponentMessageGetValueRequest{
		Key: k,
	}, ComponentMessageTypeGetValueResponse)
	return payload.Value
}
func (c *ComponentStorage) setValue(k, v string) {
	c.io.NewThread().Send(ComponentMessageTypeStoreValueRequest, ComponentMessageSetValueRequest{
		Key:   k,
		Value: v,
	})
}

func (c *ComponentStorage) GetString(k string) string {
	return c.getValue(k)
}
func (c *ComponentStorage) SetString(k, v string) {
	c.setValue(k, v)
}
