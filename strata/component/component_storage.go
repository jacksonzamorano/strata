package component

import (
	"strconv"
	"time"

	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type ComponentStorage struct {
	io *componentipc.IO
}

func newComponentStorage(io *componentipc.IO) *ComponentStorage {
	return &ComponentStorage{io: io}
}

func (c *ComponentStorage) getValue(k string) string {
	thread := c.io.NewThread()
	payload, _ := componentipc.SendAndReceive[componentipc.ComponentMessageGetValueResponse](
		thread,
		componentipc.ComponentMessageTypeGetValueRequest,
		componentipc.ComponentMessageGetValueRequest{Key: k},
		componentipc.ComponentMessageTypeGetValueResponse,
	)
	return payload.Value
}

func (c *ComponentStorage) setValue(k, v string) {
	c.io.NewThread().Send(componentipc.ComponentMessageTypeStoreValueRequest, componentipc.ComponentMessageSetValueRequest{
		Key:   k,
		Value: v,
	})
}

func (c *ComponentStorage) GetString(k string) string {
	return c.getValue(k)
}

func (c *ComponentStorage) SetString(k, v string) error {
	c.setValue(k, v)
	return nil
}

func (c *ComponentStorage) GetInt(k string) int {
	val := c.GetString(k)
	i, e := strconv.Atoi(val)
	if e != nil {
		return 0
	}
	return i
}

func (c *ComponentStorage) SetInt(k string, v int) error {
	return c.SetString(k, strconv.Itoa(v))
}

func (c *ComponentStorage) GetFloat(k string) float64 {
	val := c.GetString(k)
	f, e := strconv.ParseFloat(val, 64)
	if e != nil {
		return 0
	}
	return f
}

func (c *ComponentStorage) GetBool(k string) bool {
	val := c.GetString(k)
	b, e := strconv.ParseBool(val)
	if e != nil {
		return false
	}
	return b
}

func (c *ComponentStorage) GetDate(k string) time.Time {
	val := c.GetString(k)
	t, e := time.Parse(time.RFC3339, val)
	if e != nil {
		return time.Time{}
	}
	return t
}
