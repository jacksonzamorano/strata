import Foundation
import Passport

@Enum
enum EventKind: String {
    case componentRegistered,
         componentReady,
         taskRegistered,
         taskStarted,
         taskFinished,
         componentFunctionStarted,
         componentFunctionFinished
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
struct EventComponentReadyPayload {
    let name = Field(.string)
    let succeeded = Field(.bool)
    let error = Field(.optional(.string))
}

@Model
struct EventTaskStartedPayload {
    let id = Field(.string)
    let name = Field(.string)
    let date = Field(.datetime)
}

@Model
struct EventTaskFinishedPayload {
    let id = Field(.string)
    let name = Field(.string)
    let date = Field(.datetime)
    let duration = Field(.double)
    let succeeded = Field(.bool)
}

@Model
struct EventComponentFunctionStartedPayload {
    let id = Field(.string)
    let component = Field(.string)
    let function = Field(.string)
    let date = Field(.datetime)
}

@Model
struct EventComponentFunctionFinishedPayload {
    let id = Field(.string)
    let component = Field(.string)
    let function = Field(.string)
    let date = Field(.datetime)
    let duration = Field(.double)
    let succeeded = Field(.bool)
    let value = Field(.string)
    let error = Field(.optional(.string))
}
