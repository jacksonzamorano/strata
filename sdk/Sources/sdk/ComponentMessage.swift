import Passport
import Foundation

@Enum
enum ComponentMessageType: String {
    case ready, initialize, execute, ret, error, log
}

@Model
struct ComponentReadyMessage {
    var name = Field(.string)
    var version = Field(.string)
}
