import Foundation
import Passport

@Enum
enum HostMessageType: String {
    case tasksList,
         componentsList,
         logEvent,
         requestHistory,
         authorizationsList,
         permissionRequest,
         pendingPermissionList,
         getTasksList,
         getComponentsList,
         getRequestHistory,
         getAuthorizationsList,
         getPendingPermissionList,
         createAuthorization,
         respondPermission
}

@Model
struct HostMessageAuthorizationCreated {
    let nickname = Field(.optional(.string))
    let secret = Field(.string)
    let source = Field(.string)
    let createdDate = Field(.datetime)
}

@Model
struct HostMessageLogEvent {
    let date = Field(.datetime)
    let channel = Field(.string)
    let kind = Field(.string)
    let namespace = Field(.optional(.string))
    let message = Field(.string)
    let payload = Field(.optional(.string))
}

@Model
struct HostMessageRequestPermission {
    let permission = Field(.model(Permission.self))
}

@Model
struct HostMessageRespondPermission {
    let approve = Field(.bool)
}

@Model
struct HostStatusTask {
    let name = Field(.string)
    let url = Field(.string)
}

@Model
struct HostStatusComponent {
    let name = Field(.string)
    let version = Field(.string)
    let isHealthy = Field(.bool)
}

@Model
struct HostStatusPendingPermission {
    let id = Field(.string)
    let permission = Field(.model(Permission.self))
}

@Model
struct HostRequestHistoryEntry {
    let id = Field(.int64)
    let succeeded = Field(.bool)
    let inputBody = Field(.string)
    let inputQuery = Field(.string)
    let inputHeaders = Field(.string)
    let output = Field(.optional(.string))
    let taskStartDate = Field(.datetime)
    let taskEndDate = Field(.datetime)
}

@Model
struct HostMessageTasksList {
    let tasks = Field(.array(.model(HostStatusTask.self)))
}

@Model
struct HostMessageComponentsList {
    let components = Field(.array(.model(HostStatusComponent.self)))
}

@Model
struct HostMessageRequestHistory {
    let requests = Field(.array(.model(HostRequestHistoryEntry.self)))
}

@Model
struct HostMessageAuthorizationsList {
    let authorizations = Field(.array(.model(HostMessageAuthorizationCreated.self)))
}

@Model
struct HostMessagePendingPermissionList {
    let pendingPermissions = Field(.array(.model(HostStatusPendingPermission.self)))
}

@Model
struct HostMessageGetTasksList {}

@Model
struct HostMessageGetComponentsList {}

@Model
struct HostMessageGetRequestHistory {}

@Model
struct HostMessageGetAuthorizationsList {}

@Model
struct HostMessageGetPendingPermissionList {}

@Model
struct HostMessageCreateAuthorization {
    let nickname = Field(.string)
}
