package strata

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/hostio"
)

type HostIOService struct {
	persistence core.PersistenceProvider
	host        *hostio.IO

	lock               sync.RWMutex
	pendingPermissions map[string]*pendingPermissionRequest
}

type pendingPermissionRequest struct {
	permission core.Permission
	waiter     chan bool
}

func newAppHostService(persistence core.PersistenceProvider, host *hostio.IO) *HostIOService {
	service := &HostIOService{
		persistence: persistence,
		host:        host,

		pendingPermissions: map[string]*pendingPermissionRequest{},
	}
	go service.listenForHostMessages()
	return service
}

func (hs *HostIOService) Done() <-chan struct{} {
	return hs.host.Done()
}

func (hs *HostIOService) Emit(typ hostio.HostMessageType, payload any) {
	hs.host.Send(typ, payload)
}

func (hs *HostIOService) Log(v string, args ...any) {
	hs.host.Send(hostio.HostMessageTypeLogEvent, hostio.HostMessageLogEvent{
		Namespace: "global",
		Message:   fmt.Sprintf(v, args...),
	})
}

func (hs *HostIOService) Container(namespace string) core.Logger {
	return &appHostContainerLogger{
		service:   hs,
		namespace: namespace,
	}
}

func (hs *HostIOService) RequestPermission(permission core.Permission) bool {
	var waiter chan bool

	permission_hash := fmt.Sprintf("%s.%s.%s", permission.Container, permission.Action, *permission.Scope)
	hs.lock.Lock()
	if existing, ok := hs.pendingPermissions[permission_hash]; ok {
		waiter = existing.waiter
	} else {
		waiter = make(chan bool, 1)
		defer func() {
			hs.lock.Lock()
			delete(hs.pendingPermissions, permission_hash)
			hs.lock.Unlock()
		}()
		go func() {
			res, _ := hostio.SendAndReceive[hostio.HostMessageRespondPermission](hs.host.NewThread(), hostio.HostMessageTypePermissionRequest, hostio.HostMessageRequestPermission{
				Permission: permission,
			}, hostio.HostMessageTypeRespondPermission)
			waiter <- res.Approve
		}()
		hs.pendingPermissions[permission_hash] = &pendingPermissionRequest{
			permission: permission,
			waiter:     waiter,
		}
	}
	hs.lock.Unlock()

	select {
	case approved := <-waiter:
		return approved
	case <-time.After(2 * time.Minute):
		return false
	case <-hs.host.Done():
		return false
	}
}

func (hs *HostIOService) closePendingPermissionRequest(id string) {
	hs.lock.Lock()
	delete(hs.pendingPermissions, id)
	hs.lock.Unlock()
}

func (hs *HostIOService) listenForHostMessages() {
	getAuthorizationsList := hostio.Receive[hostio.HostMessageGetAuthorizationsList](hs.host, hostio.HostMessageTypeGetAuthorizationsList)
	createAuthorization := hostio.Receive[hostio.HostMessageCreateAuthorization](hs.host, hostio.HostMessageTypeCreateAuthorization)

	hs.host.Send(hostio.HostMessageTypeHello, struct{}{})

	for {
		select {
		case ev := <-getAuthorizationsList:
			if ev.Error {
				continue
			}
			hs.sendAuthorizationsList()
		case ev := <-createAuthorization:
			if ev.Error {
				continue
			}
			hs.handleCreateAuthorization(ev)
		}
	}
}

func (hs *HostIOService) handleCreateAuthorization(ev hostio.ReceivedEvent[hostio.HostMessageCreateAuthorization]) {
	nickname := strings.TrimSpace(ev.Payload.Nickname)
	if len(nickname) == 0 {
		hs.Log("Invalid payload: host requires a name")
		return
	}

	hs.persistence.Authorization.NewAuthorization("Host", nickname)
}

func (hs *HostIOService) sendAuthorizationsList() {
	authorizations := hs.persistence.Authorization.GetAuthorizations()
	statusAuthorizations := make([]hostio.HostMessageAuthorizationCreated, 0, len(authorizations))
	for i := range authorizations {
		auth := authorizations[i]
		statusAuthorizations = append(statusAuthorizations, hostio.HostMessageAuthorizationCreated{
			Nickname:    auth.Nickname,
			Secret:      auth.Secret,
			Source:      auth.Source,
			CreatedDate: auth.CreatedDate,
		})
	}
	hs.host.Send(hostio.HostMessageTypeAuthorizationsList, hostio.HostMessageAuthorizationsList{
		Authorizations: statusAuthorizations,
	})
}

type appHostContainerLogger struct {
	service   *HostIOService
	namespace string
}

func (l *appHostContainerLogger) Log(v string, args ...any) {
	l.service.host.Send(hostio.HostMessageTypeLogEvent, hostio.HostMessageLogEvent{
		Namespace: l.namespace,
		Message:   fmt.Sprintf(v, args...),
	})
}
