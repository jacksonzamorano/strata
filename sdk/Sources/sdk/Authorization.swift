import Foundation
import Passport

@Record(type: .table("authorization"))
struct Authorization {
    let id = Field(.int64, .primaryKey)

    let createdDate = Field(.datetime, .defaultValue("CURRENT_TIMESTAMP"))
    let lastUsedDate = Field(.datetime, .defaultValue("CURRENT_TIMESTAMP"))

    let nickname = Field(.optional(.string))
    let secret = Field(.string)
    let active = Field(.bool, .defaultValue("1"))
    let source = Field(.string)

    static let createAuthorization = insert(\.nickname, \.secret, \.source)

    @Argument
    struct GetAuth {
        let _secret: DataType = .string
    }
    static let useAuthorization = update(with: GetAuth.self) { q in
        q.set("\(\.lastUsedDate) = CURRENT_TIMESTAMP")
        q.filter("\(\.secret) = \(\._secret)")
        q.one()
    }

    static let deleteAuthorization = delete(
        with: GetAuth.self,
        { q in
            q.filter("\(\.secret) = \(\._secret)")
        })

    @Argument
    struct NoArguments {}
    static let getAuthorizations = select(with: NoArguments.self) { q in
        q.many()
    }
}
