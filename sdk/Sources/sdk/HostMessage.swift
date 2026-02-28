import Foundation
import Passport

@Enum
enum HostMessageType: String {
    case hello,
         helloAck,
         subscribeLogs,
         subscribeLogsAck,
         authorizationCreate,
         authorizationCreated,
         eventReceived,
         permissionRequest,
         permissionReplied,
         error
}

@Model
struct HostMessage {
    let id = Field(.string)
    let type = Field(.value(HostMessageType.self))
    let payload = Field(.model(HostMessagePayload.self))
}

@Model
struct HostMessagePayload {
    let hello = Field(.optional(.model(HostMessageHello.self)))
    let helloAck = Field(.optional(.model(HostMessageHelloAck.self)))

    let subscribeLogs = Field(.optional(.model(HostMessageSubscribeLogs.self)))
    let subscribeLogsAck = Field(.optional(.model(HostMessageSubscribeLogsAck.self)))

    let authorizationCreate = Field(.optional(.model(HostMessageAuthorizationCreate.self)))
    let authorizationCreated = Field(.optional(.model(HostMessageAuthorizationCreated.self)))
    
    let requestPermission = Field(.optional(.model(HostMessagePermissionResponded.self)))
    let permissionResponse = Field(.optional(.model(HostMessagePermissionResponded.self)))

    let eventReceived = Field(.optional(.model(HostMessageEventReceived.self)))
    let error = Field(.optional(.model(HostMessageError.self)))
}

@Model
struct HostMessageHello {
    let protocolVersion = Field(.string)
    let clientName = Field(.string)
}

@Model
struct HostMessageHelloAck {
    let protocolVersion = Field(.string)
    let serverName = Field(.string)
}

@Model
struct HostMessageSubscribeLogs {
    let tail = Field(.int64)
}

@Model
struct HostMessageSubscribeLogsAck {
    let replayCount = Field(.int64)
}

@Model
struct HostMessageAuthorizationCreate {
    let nickname = Field(.string)
}

@Model
struct HostMessageAuthorizationCreated {
    let nickname = Field(.optional(.string))
    let secret = Field(.string)
    let source = Field(.string)
    let createdDate = Field(.datetime)
}

@Model
struct HostMessageEventReceived {
    let date = Field(.datetime)
    let channel = Field(.string)
    let kind = Field(.string)
    let namespace = Field(.optional(.string))
    let message = Field(.string)
    let payload = Field(.optional(.string))
}

@Model
struct HostMessageError {
    let code = Field(.string)
    let message = Field(.string)
}

@Model
struct HostMessageRequestPermission {
    let permission = Field(.model(Permission.self))
}

@Model
struct HostMessagePermissionResponded {
    let approve = Field(.bool)
}
