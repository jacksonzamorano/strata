import Passport
import Foundation

@Enum
enum PermissionAction: String {
    case readFile
}

@Model
struct Permission {
    let container = Field(.string)
    let action = Field(.value(PermissionAction.self))
    let scope = Field(.optional(.string))
}
