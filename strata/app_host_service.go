package strata

import (
	"fmt"
	"slices"
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

func (hs *HostIOService) sendMessage(id string, typ hostio.HostMessageType, payload any) bool {
	if len(id) == 0 {
		id = makeId()
	}

	msg, err := hostio.NewHostMessage(id, typ, payload)
	if err != nil {
		return false
	}
	return hs.host.Send(msg)
}

func (hs *HostIOService) Emit(typ hostio.HostMessageType, payload any) {
	hs.sendMessage("", typ, payload)
}

func (hs *HostIOService) emitLog(kind, namespace, message string) {
	hs.sendMessage("", hostio.HostMessageTypeLogEvent, hostio.HostMessageLogEvent{
		Kind:      kind,
		Namespace: namespace,
		Message:   message,
	})
}

func (hs *HostIOService) Info(v string, args ...any) {
	hs.emitLog("info", "", fmt.Sprintf(v, args...))
}

func (hs *HostIOService) Container(namespace string) core.Logger {
	return &appHostContainerLogger{
		service:   hs,
		namespace: namespace,
	}
}

func (hs *HostIOService) RequestPermission(permission core.Permission) bool {
	requestID := makeId()
	waiter := make(chan bool, 1)

	hs.lock.Lock()
	hs.pendingPermissions[requestID] = &pendingPermissionRequest{
		permission: permission,
		waiter:     waiter,
	}
	hs.lock.Unlock()

	hs.sendMessage(requestID, hostio.HostMessageTypePermissionRequest, hostio.HostMessageRequestPermission{
		Permission: permission,
	})

	select {
	case approved := <-waiter:
		return approved
	case <-time.After(2 * time.Minute):
		hs.closePendingPermissionRequest(requestID)
		return false
	case <-hs.host.Done():
		hs.closePendingPermissionRequest(requestID)
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
	respondPermission := hostio.Receive[hostio.HostMessageRespondPermission](hs.host, hostio.HostMessageTypeRespondPermission)

	for {
		select {
		case ev := <-getAuthorizationsList:
			if ev.Error {
				return
			}
			hs.sendAuthorizationsList(ev.Message.Id)
		case ev := <-createAuthorization:
			if ev.Error {
				return
			}
			hs.handleCreateAuthorization(ev)
		case ev := <-respondPermission:
			if ev.Error {
				return
			}
			hs.handleRespondPermission(ev)
		}
	}
}

func (hs *HostIOService) handleCreateAuthorization(ev hostio.ReceivedEvent[hostio.HostMessageCreateAuthorization]) {
	nickname := strings.TrimSpace(ev.Payload.Nickname)
	if len(nickname) == 0 {
		hs.emitLog("invalidPayload", "host", "createAuthorization requires a nickname.")
		return
	}

	hs.persistence.Authorization.NewAuthorization("Host", nickname)
}

func (hs *HostIOService) handleRespondPermission(ev hostio.ReceivedEvent[hostio.HostMessageRespondPermission]) {
	hs.lock.Lock()
	pending, ok := hs.pendingPermissions[ev.Message.Id]
	if ok {
		delete(hs.pendingPermissions, ev.Message.Id)
	}
	hs.lock.Unlock()

	if !ok || pending == nil {
		hs.emitLog("unknownPermissionRequest", "host", "respondPermission referenced an unknown request id.")
		return
	}

	select {
	case pending.waiter <- ev.Payload.Approve:
	default:
	}
}

func (hs *HostIOService) readAuthorizations() []hostio.HostMessageAuthorizationCreated {
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

	slices.SortFunc(statusAuthorizations, func(a, b hostio.HostMessageAuthorizationCreated) int {
		if a.CreatedDate.Equal(b.CreatedDate) {
			return 0
		}
		if a.CreatedDate.After(b.CreatedDate) {
			return -1
		}
		return 1
	})
	return statusAuthorizations
}

func (hs *HostIOService) sendAuthorizationsList(id string) {
	hs.sendMessage(id, hostio.HostMessageTypeAuthorizationsList, hostio.HostMessageAuthorizationsList{
		Authorizations: hs.readAuthorizations(),
	})
}

type appHostContainerLogger struct {
	service   *HostIOService
	namespace string
}

func (l *appHostContainerLogger) Log(v string, args ...any) {
	l.service.emitLog("container", l.namespace, fmt.Sprintf(v, args...))
}
