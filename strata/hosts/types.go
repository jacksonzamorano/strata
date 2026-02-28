package hosts

import "github.com/jacksonzamorano/tasks/strata/core"

type Message = core.HostMessage
type MessageType = core.HostMessageType
type ReceivedEvent = core.HostReceivedEvent

const (
	MessageTypeTasksList                MessageType = core.HostMessageTypeTasksList
	MessageTypeComponentsList           MessageType = core.HostMessageTypeComponentsList
	MessageTypeLogEvent                 MessageType = core.HostMessageTypeLogEvent
	MessageTypeRequestHistory           MessageType = core.HostMessageTypeRequestHistory
	MessageTypeAuthorizationsList       MessageType = core.HostMessageTypeAuthorizationsList
	MessageTypePermissionRequest        MessageType = core.HostMessageTypePermissionRequest
	MessageTypePendingPermissionList    MessageType = core.HostMessageTypePendingPermissionList
	MessageTypeGetTasksList             MessageType = core.HostMessageTypeGetTasksList
	MessageTypeGetComponentsList        MessageType = core.HostMessageTypeGetComponentsList
	MessageTypeGetRequestHistory        MessageType = core.HostMessageTypeGetRequestHistory
	MessageTypeGetAuthorizationsList    MessageType = core.HostMessageTypeGetAuthorizationsList
	MessageTypeGetPendingPermissionList MessageType = core.HostMessageTypeGetPendingPermissionList
	MessageTypeCreateAuthorization      MessageType = core.HostMessageTypeCreateAuthorization
	MessageTypeRespondPermission        MessageType = core.HostMessageTypeRespondPermission
)

type MessageAuthorizationCreated = core.HostMessageAuthorizationCreated
type MessageLogEvent = core.HostMessageLogEvent
type MessageRequestPermission = core.HostMessageRequestPermission
type MessageRespondPermission = core.HostMessageRespondPermission
type StatusTask = core.HostStatusTask
type StatusComponent = core.HostStatusComponent
type StatusPendingPermission = core.HostStatusPendingPermission
type RequestHistoryEntry = core.HostRequestHistoryEntry
type MessageTasksList = core.HostMessageTasksList
type MessageComponentsList = core.HostMessageComponentsList
type MessageRequestHistory = core.HostMessageRequestHistory
type MessageAuthorizationsList = core.HostMessageAuthorizationsList
type MessagePendingPermissionList = core.HostMessagePendingPermissionList
type MessageGetTasksList = core.HostMessageGetTasksList
type MessageGetComponentsList = core.HostMessageGetComponentsList
type MessageGetRequestHistory = core.HostMessageGetRequestHistory
type MessageGetAuthorizationsList = core.HostMessageGetAuthorizationsList
type MessageGetPendingPermissionList = core.HostMessageGetPendingPermissionList
type MessageCreateAuthorization = core.HostMessageCreateAuthorization
