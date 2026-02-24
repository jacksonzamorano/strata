import Passport
import Foundation

@Enum
enum ComponentMessageType: String {
    case hello,
         setup,
         ready,
         execute,
         ret,
         storeValueRequest,
         storeValueResponse,
         getValueRequest,
         getValueResponse,
         log
}
