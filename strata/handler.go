package strata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/hostio"
)

func (as *AppState) handle(r *http.Request, task Task, container *Container) (*RequestInfo, *TaskResult) {
	requestInfo := RequestInfo{
		HasBody: false,
		Headers: r.Header,
		Query:   r.URL.Query(),
	}

	if authSec, ok := r.Header["Authorization"]; ok && len(authSec) > 0 {
		requestInfo.Authorization = as.persistence.Authorization.GetAuthorization(authSec[0])
	}

	if r.Body != nil && r.Body != http.NoBody {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return &requestInfo, &TaskResult{
				Success:    false,
				StatusCode: 400,
				Result:     "Could not read request body.",
			}
		}
		requestInfo.Body = b
		requestInfo.HasBody = true
	}

	return &requestInfo, task.Function(&requestInfo, container)
}

func (as *AppState) handler(ar Task, container *Container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		id := makeId()
		as.host.Emit(hostio.HostMessageTypeTaskTriggered, hostio.HostMessageTaskTriggered{
			Id:   id,
			Name: ar.Name,
			Date: start,
		})

		// Build request & dispatch
		requestInfo, response := as.handle(r, ar, container)

		// If we didn't catch a response,
		// make sure we send something.
		if response == nil {
			response = &TaskResult{
				Success: false,
				Result:  "Your request could not be completed.",
			}
		}

		// Serialize response into a real body.
		var outputBody []byte
		var err error
		if str, ok := response.Result.(string); ok {
			outputBody = []byte(str)
		} else if bya, ok := response.Result.([]byte); ok {
			outputBody = bya
		} else {
			outputBody, err = json.Marshal(response.Result)
		}
		if err != nil {
			genericErr := fmt.Sprintf("An error occured and your request could not be completed: %s", err.Error())
			w.Write([]byte(genericErr))
			return
		}

		// Derive status code if not provided.
		if response.StatusCode == 0 && response.Success {
			response.StatusCode = 200
		} else if response.StatusCode == 0 && !response.Success {
			response.StatusCode = 500
		}

		// Write body
		w.WriteHeader(response.StatusCode)
		w.Write(outputBody)

		// Save task run
		end := time.Now()
		as.persistence.TaskHistoryStorage.SaveTaskRun(core.TaskHistoryEntry{
			Successful:   response.Success,
			Input:        requestInfo.Body,
			Output:       outputBody,
			QueryParams:  requestInfo.Query,
			HeaderParams: requestInfo.Headers,
			Start:        start,
			End:          end,
		})
	}
}
