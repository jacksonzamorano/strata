package component

import "github.com/jacksonzamorano/strata/internal/componentipc"

type ComponentExecuteResponse struct {
	Ok     bool
	Output string
	Error  string
}

func (c *ComponentContainer) Run(program string, args ...string) ComponentExecuteResponse {
	thread := c.channel.NewThread()
	result, _ := componentipc.SendAndReceive[componentipc.ComponentMessageExecuteProgramResponse](
		thread,
		componentipc.ComponentMessageTypeExecuteProgramRequest,
		componentipc.ComponentMessageExecuteProgramRequest{
			Program:   program,
			Arguments: args,
		},
		componentipc.ComponentMessageTypeExecuteProgramResponse,
	)
	return ComponentExecuteResponse{
		Ok:     result.Ok,
		Output: result.Output,
		Error:  result.Error,
	}
}

func (c *ComponentContainer) OpenUrl(url string) bool {
	thread := c.channel.NewThread()
	result, _ := componentipc.SendAndReceive[componentipc.ComponentMessageLaunchUrlResponse](
		thread,
		componentipc.ComponentMessageTypeLaunchUrlRequest,
		componentipc.ComponentMessageLaunchUrlRequest{
			Url: url,
		},
		componentipc.ComponentMessageTypeLaunchUrlResponse,
	)
	return result.Completed
}
