package component

import (
	"github.com/jacksonzamorano/strata/internal/componentipc"
)

func (c *ComponentContainer) RequestOauth(url, callback string) (string, bool) {
	thread := c.channel.NewThread()
	response, _ := componentipc.SendAndReceive[componentipc.ComponentMessageCompleteOauthAuthentication](
		thread,
		componentipc.ComponentMessageTypeRequestOauthAuthentication,
		componentipc.ComponentMessageRequestOauthAuthentication{
			Url:      url,
			Callback: callback,
		},
		componentipc.ComponentMessageTypeCompleteOauthAuthentication,
	)
	if len(response.Url) == 0 {
		return "", false
	}
	return response.Url, true
}

func (c *ComponentContainer) RequestSecret(prompt string) (string, bool) {
	thread := c.channel.NewThread()
	response, _ := componentipc.SendAndReceive[componentipc.ComponentMessageCompleteSecretAuthentication](
		thread,
		componentipc.ComponentMessageTypeRequestSecretAuthentication,
		componentipc.ComponentMessageRequestSecretAuthentication{
			Prompt: prompt,
		},
		componentipc.ComponentMessageTypeCompleteSecretAuthentication,
	)
	if len(response.Secret) == 0 {
		return "", false
	}
	return response.Secret, true
}
