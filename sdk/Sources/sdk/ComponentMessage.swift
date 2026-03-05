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
         getValueRequest,
         getValueResponse,
         storeKeychainRequest,
         getKeychainRequest,
         getKeychainResponse,
         sendTrigger,
         log
}
