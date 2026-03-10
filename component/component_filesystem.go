package component

import (
	"os"

	"github.com/jacksonzamorano/strata/internal/componentipc"
)

func (c *ComponentContainer) ReadFile(name string) ([]byte, bool) {
	t := c.channel.NewThread()
	r, _ := componentipc.SendAndReceive[componentipc.ComponentMessageReadFileResponse](
		t,
		componentipc.ComponentMessageTypeReadFileRequest,
		componentipc.ComponentMessageReadFileRequest{
			Path: name,
		},
		componentipc.ComponentMessageTypeReadFileResponse,
	)
	if !r.Succeeded {
		return nil, false
	}
	if len(r.Path) > 0 {
		contents, err := os.ReadFile(name)
		if err != nil {
			c.Logger.Log("ReadFile: Could not read linked file: %s", err.Error())
			return nil, false
		}
		return contents, true
	}
	return r.Contents, true
}
