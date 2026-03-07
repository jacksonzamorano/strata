import Foundation
import Passport

let sql = SQLBuilder(SQLite())

let sourceFile = URL(filePath: #filePath)
let sdkRoot = sourceFile
    .deletingLastPathComponent()
    .deletingLastPathComponent()
    .deletingLastPathComponent()
let projectRoot = sdkRoot.deletingLastPathComponent()

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
        root: projectRoot.appending(path: "core"),
        fileStrategy: .monolithic,
        generateRecords: .asRecords,
        generateModels: true
    )
}
.sql(sql, rootDirectory: projectRoot)
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
        root: projectRoot.appending(path: "hostio"),
        fileStrategy: .monolithic,
        generateRecords: .asRecords,
        generateModels: true
    )
}
.build()

try! Schema("messagetypes") {
    ComponentMessageType.self
} routes: {

}
.output(Go(sqlBuilder: sql, config: ipcGoConfig)) {
    CodeBuilderConfiguration(
        root: projectRoot.appending(path: "internal/componentipc"),
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
    HostMessageRequestOauth.self
    HostMessageCompleteOauth.self
    HostMessageRequestSecret.self
    HostMessageCompleteSecret.self
} routes: {

}
.output(Go(sqlBuilder: sql, config: hostGoConfig)) {
    CodeBuilderConfiguration(
        root: projectRoot.appending(path: "hostio"),
        fileStrategy: .monolithic,
        generateRecords: .none,
        generateModels: true
    )
}
.build()
