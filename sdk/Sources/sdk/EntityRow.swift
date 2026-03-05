import Foundation
import Passport

@Record(type: .table("entity_row"))
struct EntityRow {
    let id = Field(.int64, .primaryKey)

    let createdDate = Field(.datetime, .defaultValue("CURRENT_TIMESTAMP"))
    let modifiedDate = Field(.datetime, .defaultValue("CURRENT_TIMESTAMP"))

    let namespace = Field(.string)
    let kind = Field(.string)
    let value = Field(.string)

    static let createEntityRow = insert(\.namespace, \.kind, \.value)

    @Argument
    struct UpdateArgs {
        let _namespace: DataType = .string
        let _kind: DataType = .string
        let _id: DataType = .int64
        let _value: DataType = .string
    }
    static let updateEntityRow = update(with: UpdateArgs.self) { q in
        q.filter(
            "\(\.namespace) = \(\._namespace) AND \(\.kind) = \(\._kind) AND \(\.id) = \(\._id)")
        q.set("\(\.value) = \(\._value)")
    }

    @Argument
    struct LookupArgs {
        let _namespace: DataType = .string
        let _kind: DataType = .string
        let _id: DataType = .int64
    }
    static let getEntityRow = select(with: LookupArgs.self) { q in
        q.filter(
            "\(\.namespace) = \(\._namespace) AND \(\.kind) = \(\._kind) AND \(\.id) = \(\._id)")
        q.one()
    }
    static let deleteEntityRow = delete(with: LookupArgs.self) { q in
        q.filter(
            "\(\.namespace) = \(\._namespace) AND \(\.kind) = \(\._kind) AND \(\.id) = \(\._id)")
    }

    @Argument
    struct Namespace {
        let _namespace: DataType = .string
        let _kind: DataType = .string
    }
    static let getInNamespace = select(with: Namespace.self) { q in
        q.filter("\(\.namespace) = \(\._namespace) AND \(\.kind) = \(\._kind)")
        q.many()
    }
}
