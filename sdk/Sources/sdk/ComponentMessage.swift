import Passport
import Foundation

@Enum
enum MessageType: String {
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
         log
}
