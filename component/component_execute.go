package component

import "github.com/jacksonzamorano/strata/internal/componentipc"

type ComponentExecuteResponse struct {
	Ok     bool
	Output string
	Code   int
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
		Code:   result.Code,
		Error:  result.Error,
	}
}

func (c *ComponentContainer) RunInDirectory(wd string, program string, args ...string) ComponentExecuteResponse {
	thread := c.channel.NewThread()
	result, _ := componentipc.SendAndReceive[componentipc.ComponentMessageExecuteProgramResponse](
		thread,
		componentipc.ComponentMessageTypeExecuteProgramRequest,
		componentipc.ComponentMessageExecuteProgramRequest{
			WorkingDirectory: wd,
			Program:          program,
			Arguments:        args,
		},
		componentipc.ComponentMessageTypeExecuteProgramResponse,
	)
	return ComponentExecuteResponse{
		Ok:     result.Ok,
		Output: result.Output,
		Code:   result.Code,
		Error:  result.Error,
	}
}

type ComponentDaemonConfig struct {
	WorkingDirectory string
	Program          string
	Args             []string
	Exited           func(ComponentExecuteResponse)
}

func (c *ComponentContainer) StartDaemonInDirectory(cfg ComponentDaemonConfig) ComponentExecuteResponse {
	thread := c.channel.NewThread()
	start, _ := componentipc.SendAndReceive[componentipc.ComponentMessageExecuteProgramStartedResponse](
		thread,
		componentipc.ComponentMessageTypeExecuteProgramRequest,
		componentipc.ComponentMessageExecuteProgramRequest{
			WorkingDirectory: cfg.WorkingDirectory,
			Program:          cfg.Program,
			Arguments:        cfg.Args,
			Background:       true,
		},
		componentipc.ComponentMessageTypeExecuteProgramStartedResponse,
	)

	go func() {
		result, _ := componentipc.WaitFor[componentipc.ComponentMessageExecuteProgramResponse](
			thread,
			componentipc.ComponentMessageTypeExecuteProgramResponse,
		)

		cfg.Exited(ComponentExecuteResponse{
			Ok:     result.Ok,
			Output: result.Output,
			Code:   result.Code,
			Error:  result.Error,
		})
	}()
	return ComponentExecuteResponse{
		Ok:    start.Ok,
		Error: start.Error,
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
