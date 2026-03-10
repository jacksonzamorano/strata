import Foundation
import Passport

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
        requestOauthAuthentication,
        completeOauthAuthentication,
        requestSecretAuthentication,
        completeSecretAuthentication,
        executeProgramRequest,
        executeProgramResponse,
        launchUrlRequest,
        launchUrlResponse,
        readFileRequest,
        readFileResponse,
        log
}
