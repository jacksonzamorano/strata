package strata

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jacksonzamorano/tasks/strata/core"
	hostbus "github.com/jacksonzamorano/tasks/strata/internal/hosts"
)

type appHostRuntimeState struct {
	lock sync.RWMutex

	tasks              map[string]core.HostStatusTask
	components         map[string]core.HostStatusComponent
	pendingPermissions map[string]*pendingPermissionRequest
}

type appHostService struct {
	persistence core.PersistenceProvider
	coordinator *hostbus.HostBusCoordinator

	messageID atomic.Uint64
	state     appHostRuntimeState
}

type pendingPermissionRequest struct {
	permission core.Permission
	waiter     chan bool
}

func newAppHostService(persistence core.PersistenceProvider, coordinator *hostbus.HostBusCoordinator) *appHostService {
	service := &appHostService{
		persistence: persistence,
		coordinator: coordinator,
		state: appHostRuntimeState{
			tasks:              map[string]core.HostStatusTask{},
			components:         map[string]core.HostStatusComponent{},
			pendingPermissions: map[string]*pendingPermissionRequest{},
		},
	}
	go service.listenForHostMessages()
	return service
}

func (hs *appHostService) nextMessageID() string {
	return fmt.Sprintf("%d", hs.messageID.Add(1))
}

func (hs *appHostService) sendMessage(id string, typ core.HostMessageType, payload any) bool {
	messageID := strings.TrimSpace(id)
	if len(messageID) == 0 {
		messageID = hs.nextMessageID()
	}

	msg, err := core.NewHostMessage(messageID, typ, payload)
	if err != nil {
		return false
	}
	return hs.coordinator.Send(msg)
}

func (hs *appHostService) emitLog(channel, kind string, namespace *string, message string, payload *string) {
	hs.sendMessage("", core.HostMessageTypeLogEvent, core.HostMessageLogEvent{
		Date:      time.Now(),
		Channel:   channel,
		Kind:      kind,
		Namespace: namespace,
		Message:   message,
		Payload:   payload,
	})
}

func (hs *appHostService) Info(v string, args ...any) {
	hs.emitLog("log", "info", nil, fmt.Sprintf(v, args...), nil)
}

func (hs *appHostService) Event(ev core.EventKind, payload any) {
	hs.updateStatusFromEvent(ev, payload)

	encoded, _ := json.Marshal(payload)
	payloadString := string(encoded)
	kind := string(ev)
	msg := fmt.Sprintf("Event(%s)", kind)
	hs.emitLog("event", kind, nil, msg, &payloadString)
}

func (hs *appHostService) Container(namespace string) core.Logger {
	return &appHostContainerLogger{
		service:   hs,
		namespace: namespace,
	}
}

func (hs *appHostService) RequestPermission(permission core.Permission) bool {
	requestID := hs.nextMessageID()
	waiter := make(chan bool, 1)

	hs.state.lock.Lock()
	hs.state.pendingPermissions[requestID] = &pendingPermissionRequest{
		permission: permission,
		waiter:     waiter,
	}
	hs.state.lock.Unlock()

	permissionJSON, _ := json.Marshal(permission)
	permissionPayload := string(permissionJSON)
	startMsg := fmt.Sprintf("[%s] requested permission '%s'. Awaiting approval.", permission.Container, permission.Action)
	hs.emitLog("permission", "requestPending", nil, startMsg, &permissionPayload)

	hs.sendMessage(requestID, core.HostMessageTypePermissionRequest, core.HostMessageRequestPermission{
		Permission: permission,
	})

	select {
	case approved := <-waiter:
		decision := "denied"
		if approved {
			decision = "approved"
		}
		endMsg := fmt.Sprintf("[%s] permission '%s' %s.", permission.Container, permission.Action, decision)
		hs.emitLog("permission", "requestResolved", nil, endMsg, &permissionPayload)
		return approved
	case <-time.After(2 * time.Minute):
		hs.state.lock.Lock()
		delete(hs.state.pendingPermissions, requestID)
		hs.state.lock.Unlock()

		timeoutMsg := fmt.Sprintf("[%s] permission '%s' timed out.", permission.Container, permission.Action)
		hs.emitLog("permission", "requestTimedOut", nil, timeoutMsg, &permissionPayload)
		return false
	}
}

func (hs *appHostService) listenForHostMessages() {
	getTasksList := hostbus.Receive[core.HostMessageGetTasksList](hs.coordinator, core.HostMessageTypeGetTasksList)
	getComponentsList := hostbus.Receive[core.HostMessageGetComponentsList](hs.coordinator, core.HostMessageTypeGetComponentsList)
	getRequestHistory := hostbus.Receive[core.HostMessageGetRequestHistory](hs.coordinator, core.HostMessageTypeGetRequestHistory)
	getAuthorizationsList := hostbus.Receive[core.HostMessageGetAuthorizationsList](hs.coordinator, core.HostMessageTypeGetAuthorizationsList)
	getPendingPermissionList := hostbus.Receive[core.HostMessageGetPendingPermissionList](hs.coordinator, core.HostMessageTypeGetPendingPermissionList)
	createAuthorization := hostbus.Receive[core.HostMessageCreateAuthorization](hs.coordinator, core.HostMessageTypeCreateAuthorization)
	respondPermission := hostbus.Receive[core.HostMessageRespondPermission](hs.coordinator, core.HostMessageTypeRespondPermission)

	for {
		select {
		case ev := <-getTasksList:
			if ev.Error {
				return
			}
			hs.sendTasksList(ev.Message.Id)
		case ev := <-getComponentsList:
			if ev.Error {
				return
			}
			hs.sendComponentsList(ev.Message.Id)
		case ev := <-getRequestHistory:
			if ev.Error {
				return
			}
			hs.sendRequestHistory(ev.Message.Id)
		case ev := <-getAuthorizationsList:
			if ev.Error {
				return
			}
			hs.sendAuthorizationsList(ev.Message.Id)
		case ev := <-getPendingPermissionList:
			if ev.Error {
				return
			}
			hs.sendPendingPermissionList(ev.Message.Id)
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

func (hs *appHostService) handleCreateAuthorization(ev hostbus.ReceivedEvent[core.HostMessageCreateAuthorization]) {
	nickname := strings.TrimSpace(ev.Payload.Nickname)
	if len(nickname) == 0 {
		hs.emitLog("host", "invalidPayload", nil, "createAuthorization requires a nickname.", nil)
		return
	}

	hs.persistence.Authorization.NewAuthorization("Host", nickname)
	hs.sendAuthorizationsList(ev.Message.Id)
}

func (hs *appHostService) handleRespondPermission(ev hostbus.ReceivedEvent[core.HostMessageRespondPermission]) {
	hs.state.lock.Lock()
	pending, ok := hs.state.pendingPermissions[ev.Message.Id]
	if ok {
		delete(hs.state.pendingPermissions, ev.Message.Id)
	}
	hs.state.lock.Unlock()

	if !ok || pending == nil {
		hs.emitLog("host", "unknownPermissionRequest", nil, "respondPermission referenced an unknown request id.", nil)
		return
	}

	select {
	case pending.waiter <- ev.Payload.Approve:
	default:
	}
}

func (hs *appHostService) updateStatusFromEvent(ev core.EventKind, payload any) {
	switch ev {
	case core.EventKindTaskRegistered:
		hs.updateTaskStatus(payload)
		hs.sendTasksList("")
	case core.EventKindComponentRegistered:
		hs.updateComponentRegisteredStatus(payload)
		hs.sendComponentsList("")
	case core.EventKindComponentReady:
		hs.updateComponentReadyStatus(payload)
		hs.sendComponentsList("")
	}
}

func (hs *appHostService) updateTaskStatus(payload any) {
	var task core.EventTaskRegisterPayload
	switch p := payload.(type) {
	case core.EventTaskRegisterPayload:
		task = p
	case *core.EventTaskRegisterPayload:
		if p == nil {
			return
		}
		task = *p
	default:
		return
	}

	name := strings.TrimSpace(task.Name)
	url := strings.TrimSpace(task.Url)
	if len(name) == 0 || len(url) == 0 {
		return
	}

	hs.state.lock.Lock()
	hs.state.tasks[name] = core.HostStatusTask{
		Name: name,
		Url:  url,
	}
	hs.state.lock.Unlock()
}

func (hs *appHostService) updateComponentRegisteredStatus(payload any) {
	var component core.EventComponentRegisteredPayload
	switch p := payload.(type) {
	case core.EventComponentRegisteredPayload:
		component = p
	case *core.EventComponentRegisteredPayload:
		if p == nil {
			return
		}
		component = *p
	default:
		return
	}

	name := strings.TrimSpace(component.Name)
	if len(name) == 0 {
		return
	}

	hs.state.lock.Lock()
	current, ok := hs.state.components[name]
	if !ok {
		current = core.HostStatusComponent{
			Name:      name,
			Version:   "unknown",
			IsHealthy: false,
		}
	}
	version := strings.TrimSpace(component.Version)
	if len(version) > 0 {
		current.Version = version
	}
	current.IsHealthy = component.Suceeded
	hs.state.components[name] = current
	hs.state.lock.Unlock()
}

func (hs *appHostService) updateComponentReadyStatus(payload any) {
	var component core.EventComponentReadyPayload
	switch p := payload.(type) {
	case core.EventComponentReadyPayload:
		component = p
	case *core.EventComponentReadyPayload:
		if p == nil {
			return
		}
		component = *p
	default:
		return
	}

	name := strings.TrimSpace(component.Name)
	if len(name) == 0 {
		return
	}

	hs.state.lock.Lock()
	current, ok := hs.state.components[name]
	if !ok {
		current = core.HostStatusComponent{
			Name:      name,
			Version:   "unknown",
			IsHealthy: false,
		}
	}
	current.IsHealthy = component.Succeeded
	hs.state.components[name] = current
	hs.state.lock.Unlock()
}

func (hs *appHostService) readTaskStatuses() []core.HostStatusTask {
	hs.state.lock.RLock()
	statuses := make([]core.HostStatusTask, 0, len(hs.state.tasks))
	for _, status := range hs.state.tasks {
		statuses = append(statuses, status)
	}
	hs.state.lock.RUnlock()

	slices.SortFunc(statuses, func(a, b core.HostStatusTask) int {
		if a.Name == b.Name {
			return 0
		}
		if a.Name < b.Name {
			return -1
		}
		return 1
	})
	return statuses
}

func (hs *appHostService) readComponentStatuses() []core.HostStatusComponent {
	hs.state.lock.RLock()
	statuses := make([]core.HostStatusComponent, 0, len(hs.state.components))
	for _, status := range hs.state.components {
		statuses = append(statuses, status)
	}
	hs.state.lock.RUnlock()

	slices.SortFunc(statuses, func(a, b core.HostStatusComponent) int {
		if a.Name == b.Name {
			return 0
		}
		if a.Name < b.Name {
			return -1
		}
		return 1
	})
	return statuses
}

func (hs *appHostService) readPendingPermissions() []core.HostStatusPendingPermission {
	hs.state.lock.RLock()
	pending := make([]core.HostStatusPendingPermission, 0, len(hs.state.pendingPermissions))
	for id, request := range hs.state.pendingPermissions {
		if request == nil {
			continue
		}
		pending = append(pending, core.HostStatusPendingPermission{
			Id:         id,
			Permission: request.permission,
		})
	}
	hs.state.lock.RUnlock()

	slices.SortFunc(pending, func(a, b core.HostStatusPendingPermission) int {
		if a.Id == b.Id {
			return 0
		}
		if a.Id < b.Id {
			return -1
		}
		return 1
	})
	return pending
}

func (hs *appHostService) readAuthorizations() []core.HostMessageAuthorizationCreated {
	authorizations := hs.persistence.Authorization.GetAuthorizations()
	statusAuthorizations := make([]core.HostMessageAuthorizationCreated, 0, len(authorizations))
	for i := range authorizations {
		auth := authorizations[i]
		statusAuthorizations = append(statusAuthorizations, core.HostMessageAuthorizationCreated{
			Nickname:    auth.Nickname,
			Secret:      auth.Secret,
			Source:      auth.Source,
			CreatedDate: auth.CreatedDate,
		})
	}

	slices.SortFunc(statusAuthorizations, func(a, b core.HostMessageAuthorizationCreated) int {
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

func (hs *appHostService) readRequestHistory() []core.HostRequestHistoryEntry {
	runs := hs.persistence.TaskHistoryStorage.GetRecentTaskRuns(200)
	out := make([]core.HostRequestHistoryEntry, 0, len(runs))
	for i := range runs {
		run := runs[i]
		out = append(out, core.HostRequestHistoryEntry{
			Id:            run.Id,
			Succeeded:     run.Succeeded,
			InputBody:     run.InputBody,
			InputQuery:    run.InputQuery,
			InputHeaders:  run.InputHeaders,
			Output:        run.Output,
			TaskStartDate: run.TaskStartDate,
			TaskEndDate:   run.TaskEndDate,
		})
	}
	return out
}

func (hs *appHostService) sendTasksList(id string) {
	hs.sendMessage(id, core.HostMessageTypeTasksList, core.HostMessageTasksList{
		Tasks: hs.readTaskStatuses(),
	})
}

func (hs *appHostService) sendComponentsList(id string) {
	hs.sendMessage(id, core.HostMessageTypeComponentsList, core.HostMessageComponentsList{
		Components: hs.readComponentStatuses(),
	})
}

func (hs *appHostService) sendRequestHistory(id string) {
	hs.sendMessage(id, core.HostMessageTypeRequestHistory, core.HostMessageRequestHistory{
		Requests: hs.readRequestHistory(),
	})
}

func (hs *appHostService) sendAuthorizationsList(id string) {
	hs.sendMessage(id, core.HostMessageTypeAuthorizationsList, core.HostMessageAuthorizationsList{
		Authorizations: hs.readAuthorizations(),
	})
}

func (hs *appHostService) sendPendingPermissionList(id string) {
	hs.sendMessage(id, core.HostMessageTypePendingPermissionList, core.HostMessagePendingPermissionList{
		PendingPermissions: hs.readPendingPermissions(),
	})
}

type appHostContainerLogger struct {
	service   *appHostService
	namespace string
}

func (l *appHostContainerLogger) Log(v string, args ...any) {
	namespace := l.namespace
	l.service.emitLog("log", "container", &namespace, fmt.Sprintf(v, args...), nil)
}
