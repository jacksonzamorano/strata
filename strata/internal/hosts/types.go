package hosts

import "github.com/jacksonzamorano/tasks/strata/core"

type Message = core.HostMessage
type MessageType = core.HostMessageType
type RawReceivedEvent = core.HostReceivedEvent

type MessageTasksList = core.HostMessageTasksList
type MessageComponentsList = core.HostMessageComponentsList
type MessageLogEvent = core.HostMessageLogEvent
type MessageRequestHistory = core.HostMessageRequestHistory
type MessageAuthorizationsList = core.HostMessageAuthorizationsList
type MessagePermissionRequest = core.HostMessageRequestPermission
type MessagePendingPermissionList = core.HostMessagePendingPermissionList

type MessageGetTasksList = core.HostMessageGetTasksList
type MessageGetComponentsList = core.HostMessageGetComponentsList
type MessageGetRequestHistory = core.HostMessageGetRequestHistory
type MessageGetAuthorizationsList = core.HostMessageGetAuthorizationsList
type MessageGetPendingPermissionList = core.HostMessageGetPendingPermissionList
type MessageCreateAuthorization = core.HostMessageCreateAuthorization
type MessageRespondPermission = core.HostMessageRespondPermission
