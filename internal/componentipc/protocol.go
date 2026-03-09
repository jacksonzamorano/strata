package componentipc

import "encoding/json"

type ComponentMessage struct {
	Id      string          `json:"id"`
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ComponentResultPayload struct {
	Success  bool
	Response json.RawMessage
	Error    string
}

type ComponentMessageHello struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ComponentMessageReady struct {
	Error string `json:"error"`
}

type ComponentMessageGetValueRequest struct {
	Key string `json:"key"`
}

type ComponentMessageGetValueResponse struct {
	Value string `json:"value"`
}

type ComponentMessageSetValueRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ComponentMessageExecute struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ComponentMessageLog struct {
	Message string `json:"message"`
}

type ComponentMessageGetKeychainRequest struct {
	Key string `json:"key"`
}

type ComponentMessageGetKeychainResponse struct {
	Value string `json:"value"`
}

type ComponentMessageSetKeychainRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ComponentMessageSendTrigger struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

type ComponentMessageRequestOauthAuthentication struct {
	Url      string `json:"url"`
	Callback string `json:"callback"`
}

type ComponentMessageCompleteOauthAuthentication struct {
	Url string `json:"url"`
}

type ComponentMessageRequestSecretAuthentication struct {
	Prompt string `json:"prompt"`
}

type ComponentMessageCompleteSecretAuthentication struct {
	Secret string `json:"secret"`
}

type ComponentMessageExecuteProgramRequest struct {
	Program          string   `json:"program"`
	Arguments        []string `json:"arguments"`
	WorkingDirectory string   `json:"working_directory"`
}

type ComponentMessageExecuteProgramResponse struct {
	Ok     bool   `json:"ok"`
	Error  string `json:"error"`
	Code   int    `json:"code"`
	Output string `json:"output"`
}

type ComponentMessageLaunchUrlRequest struct {
	Url string `json:"url"`
}

type ComponentMessageLaunchUrlResponse struct {
	Completed bool `json:"completed"`
}

type ComponentMessageSetup struct {
	StorageDir string `json:"storage_dir"`
}
