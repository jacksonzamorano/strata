import Foundation
import Passport

@Enum
enum EventKind: String {
    case taskRegistered,
         componentRegistered,
         taskTriggered
}

@Model
struct EventTaskRegisterPayload {
    let name = Field(.string)
    let url = Field(.string)
}

@Model
struct EventComponentRegisteredPayload {
    let name = Field(.string)
    let version = Field(.string)
    let suceeded = Field(.bool)
    let path = Field(.string)
    let error = Field(.optional(.string))
}

@Model
struct EventTaskTriggeredPayload {
    let id = Field(.string)
    let name = Field(.string)
    let date = Field(.datetime)
}
