import Passport
import Foundation

let sql = SQLBuilder(SQLite())

let projectRoot = URL.currentDirectory().deletingLastPathComponent()
let tasklibRoot = projectRoot.appending(path: "tasklib")

let goConfig = GoConfiguration { cfg in
    cfg.packageName = "core"
}

let ipcGoConfig = GoConfiguration { cfg in
    cfg.packageName = "componentipc"
}

try! Schema("schema") {
    TaskRun.self
    StorageRow.self
    StorageRowKeyNames.self
    EntityRow.self
    Authorization.self
    EventKind.self
    EventTaskRegisterPayload.self
    EventComponentRegisteredPayload.self
    EventComponentReadyPayload.self
    EventTaskStartedPayload.self
    EventTaskFinishedPayload.self
    EventComponentFunctionStartedPayload.self
    EventComponentFunctionFinishedPayload.self
} routes: {

}
.output(Go(sqlBuilder: sql, config: goConfig)) {
    CodeBuilderConfiguration(
        root: tasklibRoot.appending(path: "core"),
        fileStrategy: .monolithic,
        generateRecords: .asRecords,
        generateModels: true
    )
}
.sql(sql, rootDirectory: tasklibRoot)
.build()

try! Schema("messagetypes") {
    MessageType.self
} routes: {

}
.output(Go(sqlBuilder: sql, config: ipcGoConfig)) {
    CodeBuilderConfiguration(
        root: tasklibRoot.appending(path: "internal/componentipc"),
        fileStrategy: .monolithic,
        generateRecords: .none,
        generateModels: true
    )
}
.build()
