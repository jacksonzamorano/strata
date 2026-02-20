package tasklib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (as *AppState) buildContainer(namespace string) *Container {
	return &Container{
		Storage: as.storage.Container(namespace),
	}
}

func (as *AppState) handle(r *http.Request, task Task) (*RequestInfo, *TaskResult) {
	requestInfo := RequestInfo{
		HasBody: false,
		Headers: r.Header,
		Query:   r.URL.Query(),
	}

	var authorization *Authorization
	if authSec, ok := r.Header["Authorization"]; ok && len(authSec) > 0 {
		authorization = as.getAuthorization(authSec[0])
	}

	container := as.buildContainer("userspace")
	container.Authorization = authorization

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

func (as *AppState) handler(ar Task) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		as.Logger.Info("[%s] Recieved request.", ar.Name)

		// Build request & dispatch
		requestInfo, response := as.handle(r, ar)

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

		// Encode metadata
		encoded_query_params, _ := json.Marshal(requestInfo.Query)
		encoded_header_params, _ := json.Marshal(requestInfo.Headers)
		body_str := string(outputBody)

		// Save task run
		end := time.Now()
		_, err = CreateTaskRun(as.database,
			response.Success,
			string(outputBody),
			string(encoded_query_params),
			string(encoded_header_params),
			&body_str,
			start,
			end,
		)
		if err != nil {
			panic(err)
		}
	}
}
