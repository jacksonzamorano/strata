import Foundation
import Passport

@Enum
enum PermissionAction: String {
    case readFile,
        executeCommandLine,
        launchUrl
}

@Model
struct Permission {
    let container = Field(.string)
    let action = Field(.value(PermissionAction.self))
    let scope = Field(.optional(.string))
}
