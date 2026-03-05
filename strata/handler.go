package strata

// func (as *AppState) handler(ar Task, container *Container) func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		start := time.Now()
// 		id := makeId()
// 		as.host.Emit(hostio.HostMessageTypeTaskTriggered, hostio.HostMessageTaskTriggered{
// 			Id:   id,
// 			Name: ar.Name,
// 			Date: start,
// 		})

// 		// Build request & dispatch
// 		requestInfo, response := as.handle(r, ar, container)

// 		// If we didn't catch a response,
// 		// make sure we send something.
// 		if response == nil {
// 			response = &TaskResult{
// 				Success: false,
// 				Result:  "Your request could not be completed.",
// 			}
// 		}

// 		// Serialize response into a real body.
// 		var outputBody []byte
// 		var err error
// 		if str, ok := response.Result.(string); ok {
// 			outputBody = []byte(str)
// 		} else if bya, ok := response.Result.([]byte); ok {
// 			outputBody = bya
// 		} else {
// 			outputBody, err = json.Marshal(response.Result)
// 		}
// 		if err != nil {
// 			genericErr := fmt.Sprintf("An error occured and your request could not be completed: %s", err.Error())
// 			w.Write([]byte(genericErr))
// 			return
// 		}

// 		// Derive status code if not provided.
// 		if response.StatusCode == 0 && response.Success {
// 			response.StatusCode = 200
// 		} else if response.StatusCode == 0 && !response.Success {
// 			response.StatusCode = 500
// 		}

// 		// Write body
// 		w.WriteHeader(response.StatusCode)
// 		w.Write(outputBody)

// 		// Save task run
// 		end := time.Now()
// 		as.persistence.TaskHistoryStorage.SaveTaskRun(core.TaskHistoryEntry{
// 			Successful:   response.Success,
// 			Input:        requestInfo.Body,
// 			Output:       outputBody,
// 			QueryParams:  requestInfo.Query,
// 			HeaderParams: requestInfo.Headers,
// 			Start:        start,
// 			End:          end,
// 		})
// 	}
// }
