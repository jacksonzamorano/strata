import Passport
import Foundation

let sql = SQLBuilder(SQLite())

let projectRoot = URL.currentDirectory().deletingLastPathComponent()
let tasklibRoot = projectRoot.appending(path: "tasklib")

let goConfig = GoConfiguration { cfg in
    cfg.packageName = "tasklib"
}

try! Schema("schema") {
    TaskRun.self
    StorageRow.self
    StorageRowKeyNames.self
    Authorization.self
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
