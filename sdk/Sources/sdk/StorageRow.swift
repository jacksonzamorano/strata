import Foundation
import Passport

@Record(type: .table("kv_storage"))
struct StorageRow {
    let id = Field(.int64, .primaryKey)
    let createdDate = Field(.datetime, .defaultValue("CURRENT_TIMESTAMP"))
    let modifiedDate = Field(.datetime, .defaultValue("CURRENT_TIMESTAMP"))
    
    let namespace = Field(.string)
    let key = Field(.string)
    let value = Field(.string)
    
    static let createStorageRow = insert(\.namespace, \.key, \.value)
    
    @Argument
    struct UpdateStorageRow {
        let _namespace: DataType = .string
        let _key: DataType = .string
        let _value: DataType = .string
    }
    static let updateStorageRow = update(with: UpdateStorageRow.self) { q in
        q.set("\(\.modifiedDate) = CURRENT_TIMESTAMP, \(\.value) = \(\._value)")
        q.filter("\(\.namespace) = \(\._namespace) AND \(\.key) = \(\._key)")
        q.one()
    }
    
    @Argument
    struct GetStorageRow {
        let _namespace: DataType = .string
        let _key: DataType = .string
    }
    static let getStorageRow = select(with: GetStorageRow.self) { q in
        q.filter("\(\.key) = \(\._key) AND \(\.namespace) = \(\._namespace)")
        q.one()
    }
}

@Record(type: .query(StorageRow.self))
struct StorageRowKeyNames {
    let namespace = fromBase(\.namespace)
    let key = fromBase(\.key)
    
    @Argument
    struct Namespace {
        var _namespace: DataType = .string
    }
    static let getStorageRowKeyNamesInNamespace = select(with: Namespace.self) { q in
        q.filter("\(\.namespace) = \(\._namespace)")
        q.many()
    }
}
