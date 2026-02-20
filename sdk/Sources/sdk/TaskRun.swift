import Passport

@Record(type: .table("task_run"))
struct TaskRun {
    let id = Field(.int64, .primaryKey)
    
    let succeeded = Field(.bool)
    
    let inputBody = Field(.string)
    let inputQuery = Field(.string)
    let inputHeaders = Field(.string)
    
    let output = Field(.optional(.string))
    
    let taskStartDate = Field(.datetime)
    let taskEndDate = Field(.datetime)
    
    static let createTaskRun = insert(\.succeeded, \.inputBody, \.inputQuery, \.inputHeaders, \.output, \.taskStartDate, \.taskEndDate)
}
