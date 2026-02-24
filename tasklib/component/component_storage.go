package component

import "strconv"

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

func (c *ComponentStorage) GetInt(k string) int {
	val := c.GetString(k)
	i, e := strconv.Atoi(val)
	if e != nil {
		return 0
	}
	return i
}
func (c *ComponentStorage) SetInt(k string, v int) {
	c.SetString(k, strconv.Itoa(v))
}
