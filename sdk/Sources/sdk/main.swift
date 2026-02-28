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

try! Schema("schema") {
    TaskRun.self
    StorageRow.self
    StorageRowKeyNames.self
    EntityRow.self
    Authorization.self
    
    PermissionAction.self
    Permission.self
    
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
        root: strataRoot.appending(path: "core"),
        fileStrategy: .monolithic,
        generateRecords: .asRecords,
        generateModels: true
    )
}
.sql(sql, rootDirectory: strataRoot)
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

try! Schema("hostschema") {
    HostMessageType.self
    HostMessage.self
    HostMessagePayload.self
    HostMessageHello.self
    HostMessageHelloAck.self
    HostMessageSubscribeLogs.self
    HostMessageSubscribeLogsAck.self
    HostMessageAuthorizationCreate.self
    HostMessageAuthorizationCreated.self
    HostMessageEventReceived.self
    HostMessageRequestPermission.self
    HostMessagePermissionResponded.self
    HostMessageError.self
} routes: {

}
.output(Go(sqlBuilder: sql, config: goConfig)) {
    CodeBuilderConfiguration(
        root: strataRoot.appending(path: "core"),
        fileStrategy: .monolithic,
        generateRecords: .none,
        generateModels: true
    )
}
.output(TypeScript(buildIndex: true)) {
    CodeBuilderConfiguration(
        root: projectRoot.appending(path: "hosts/web/src/generated"),
        fileStrategy: .perEntity,
        generateRecords: .none,
        generateModels: true
    )
}
.build()
