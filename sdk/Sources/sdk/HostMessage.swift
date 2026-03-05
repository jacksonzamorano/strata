import Foundation
import Passport

@Enum
enum HostMessageType: String {
    case hello,
        authorizationsList,
        permissionRequest,
        logEvent,
        taskRegistered,
        componentRegistered,
        taskTriggered,
        getAuthorizationsList,
        createAuthorization,
        respondPermission,
        requestOauth,
        completeOauth,
        requestSecret,
        completeSecret
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
    let kind = Field(.string)
    let namespace = Field(.string)
    let message = Field(.string)
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
struct HostMessageTaskRegistered {
    let name = Field(.string)
    let url = Field(.string)
}

@Model
struct HostMessageComponentRegistered {
    let name = Field(.string)
    let version = Field(.string)
    let suceeded = Field(.bool)
    let path = Field(.string)
    let error = Field(.optional(.string))
}

@Model
struct HostMessageTaskTriggered {
    let id = Field(.string)
    let name = Field(.string)
    let date = Field(.datetime)
}

@Model
struct HostMessageAuthorizationsList {
    let authorizations = Field(.array(.model(HostMessageAuthorizationCreated.self)))
}

@Model
struct HostMessageGetAuthorizationsList {}

@Model
struct HostMessageCreateAuthorization {
    let nickname = Field(.string)
}

@Model
struct HostMessageRequestOauth {
    let namespace = Field(.string)
    let url = Field(.string)
    let destination = Field(.string)
}

@Model
struct HostMessageCompleteOauth {
    let url = Field(.string)
}

@Model
struct HostMessageRequestSecret {
    let namespace = Field(.string)
    let prompt = Field(.string)
}

@Model
struct HostMessageCompleteSecret {
    let secret = Field(.string)
}
