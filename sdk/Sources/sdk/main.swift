import Passport
import Foundation

let sql = SQLBuilder(SQLite())

let projectRoot = URL.currentDirectory().deletingLastPathComponent()
let tasklibRoot = projectRoot.appending(path: "tasklib")

let goConfig = GoConfiguration { cfg in
    cfg.packageName = "tasklib"
}

let componentGoConfig = GoConfiguration { cfg in
    cfg.packageName = "component"
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
} routes: {
    
}
.output(Go(sqlBuilder: sql, config: goConfig)) {
    CodeBuilderConfiguration(
        root: tasklibRoot,
        fileStrategy: .monolithic,
        generateRecords: .asRecords,
        generateModels: true
    )
}
.sql(sql, rootDirectory: tasklibRoot)
.build()

try! Schema("messagetypes") {
    ComponentMessageType.self
} routes: {
    
}
.output(Go(sqlBuilder: sql, config: componentGoConfig)) {
    CodeBuilderConfiguration(
        root: tasklibRoot.appending(path: "component"),
        fileStrategy: .monolithic,
        generateRecords: .none,
        generateModels: true
    )
}
.build()
