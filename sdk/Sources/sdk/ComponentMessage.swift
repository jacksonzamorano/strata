import Passport
import Foundation

@Enum
enum ComponentMessageType: String {
    case ready, initialize, execute, ret, error, log
}
