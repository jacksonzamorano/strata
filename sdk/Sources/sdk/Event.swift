import Foundation
import Passport

@Enum
enum EventKind: String {
    case componentRegistered, taskRegistered
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
    let error = Field(.optional(.string))
}
