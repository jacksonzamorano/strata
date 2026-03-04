import Passport
import Foundation

let sql = SQLBuilder(SQLite())

let projectRoot = URL.currentDirectory().deletingLastPathComponent()
let strataRoot = projectRoot.appending(path: "strata")

let goConfig = GoConfiguration { cfg in
    cfg.packageName = "core"
}

let ipcGoConfig = GoConfiguration { cfg in
    cfg.packageName = "componentipc"
}

let hostGoConfig = GoConfiguration { cfg in
    cfg.packageName = "hostio"
}

try! Schema("schema") {
    TaskRun.self
    StorageRow.self
    StorageRowKeyNames.self
    EntityRow.self
    Authorization.self
    
    PermissionAction.self
    Permission.self
} routes: {

}
.output(Go(sqlBuilder: sql, config: goConfig)) {
    CodeBuilderConfiguration(
        root: strataRoot.appending(path: "core"),
        fileStrategy: .monolithic,
        generateRecords: .asRecords,
        generateModels: true
    )
}
.sql(sql, rootDirectory: strataRoot)
.build()

try! Schema("hostschema") {
    EventKind.self
    EventTaskRegisterPayload.self
    EventComponentRegisteredPayload.self
    EventTaskTriggeredPayload.self
} routes: {
    
}
.output(Go(sqlBuilder: sql, config: hostGoConfig)) {
    CodeBuilderConfiguration(
        root: strataRoot.appending(path: "hostio"),
        fileStrategy: .monolithic,
        generateRecords: .asRecords,
        generateModels: true
    )
}
.build()

try! Schema("messagetypes") {
    MessageType.self
} routes: {

}
.output(Go(sqlBuilder: sql, config: ipcGoConfig)) {
    CodeBuilderConfiguration(
        root: strataRoot.appending(path: "internal/componentipc"),
        fileStrategy: .monolithic,
        generateRecords: .none,
        generateModels: true
    )
}
.build()

try! Schema("host_protocol") {
    HostMessageType.self
    HostMessageAuthorizationsList.self
    HostMessageTaskRegistered.self
    HostMessageComponentRegistered.self
    HostMessageTaskTriggered.self
    HostMessageGetAuthorizationsList.self
    HostMessageCreateAuthorization.self
    HostMessageAuthorizationCreated.self
    HostMessageLogEvent.self
    HostMessageRequestPermission.self
    HostMessageRespondPermission.self
} routes: {

}
.output(Go(sqlBuilder: sql, config: hostGoConfig)) {
    CodeBuilderConfiguration(
        root: strataRoot.appending(path: "hostio"),
        fileStrategy: .monolithic,
        generateRecords: .none,
        generateModels: true
    )
}
.build()
