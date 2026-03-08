package strata

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

type jsonRpc struct {
	JsonRPC string `json:"jsonrpc"`
	Id      int    `json:"id"`
}
type jsonRpcResult[T any] struct {
	jsonRpc
	Result T `json:"result"`
}
type jsonRpcInput struct {
	jsonRpc
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}
type mcpIntializeRequest struct {
	ProtocolVersion string        `json:"protcolVersion"`
	ClientInfo      mcpClientInfo `json:"clientInfo"`
}
type mcpClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type mcpInitialize struct {
	ProtocolVersion string          `json:"protocolVersion"`
	ServerInfo      mcpServerInfo   `json:"serverInfo"`
	Capabilities    mcpCapabilities `json:"capabilities"`
}
type mcpServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type mcpCapabilities struct {
	Tools map[string]mcpToolDef `json:"tools"`
}
type mcpToolList struct {
	Tools []mcpToolDef `json:"tools"`
}
type mcpToolDef struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema any               `json:"inputSchema"`
	Annotations mcpToolAnnotation `json:"annotations"`
}
type mcpToolAnnotation struct {
	Title           string `json:"title"`
	ReadOnlyHint    bool   `json:"readOnlyHint"`
	DestructiveHint bool   `json:"destructiveHint"`
	IdempotentHint  bool   `json:"idempotentHint"`
}
type mcpInputSchema struct {
	Type       string                         `json:"type"`
	Properties map[string]mcpInputSchemaField `json:"properties"`
	Required   []string                       `json:"required"`
}
type mcpInputSchemaField struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
type mcpToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
type mcpToolCallResult struct {
	Content           []mcpToolCallResultContent `json:"content"`
	IsError           bool                       `json:"isError"`
	StructuredContent json.RawMessage            `json:"structuredContent,omitempty"`
}
type mcpToolCallResultContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MCPToolResult struct {
	Response any
	Success  bool
	Error    string
}

type mcpToolFn = func(input json.RawMessage, ctx *TaskContext) *MCPToolResult
type mcpTool struct {
	Name        string
	Description string
	Tool        mcpToolFn
	InputSchema json.RawMessage
	Annotation  mcpToolAnnotation
}

type MCPTask struct {
	Name    string
	Version string
	Tools   map[string]mcpTool
}

func (tt *MCPTask) initialize(input *jsonRpcInput) jsonRpcResult[mcpInitialize] {
	return jsonRpcResult[mcpInitialize]{
		jsonRpc: input.jsonRpc,
		Result: mcpInitialize{
			ProtocolVersion: "2025-11-25",
			ServerInfo: mcpServerInfo{
				Name:    tt.Name,
				Version: tt.Version,
			},
			Capabilities: mcpCapabilities{
				Tools: map[string]mcpToolDef{},
			},
		},
	}
}
func (tt *MCPTask) listTools(input *jsonRpcInput) jsonRpcResult[mcpToolList] {
	tools := []mcpToolDef{}

	for i := range tt.Tools {
		tools = append(tools, mcpToolDef{
			Name:        tt.Tools[i].Name,
			Description: tt.Tools[i].Description,
			Annotations: tt.Tools[i].Annotation,
			InputSchema: tt.Tools[i].InputSchema,
		})
	}

	return jsonRpcResult[mcpToolList]{
		jsonRpc: input.jsonRpc,
		Result: mcpToolList{
			Tools: tools,
		},
	}
}
func (tt *MCPTask) callTool(input *jsonRpcInput, container *TaskContext) jsonRpcResult[mcpToolCallResult] {
	var payload mcpToolCallRequest
	json.Unmarshal(input.Params, &payload)

	var res mcpToolCallResult

	tool := tt.Tools[payload.Name]
	result := tool.Tool(payload.Arguments, container)
	if result.Success {
		if str, ok := result.Response.(string); ok {
			res.Content = []mcpToolCallResultContent{
				mcpToolCallResultContent{
					Type: "text",
					Text: str,
				},
			}
		} else {
			j, _ := json.Marshal(result.Response)
			res.Content = []mcpToolCallResultContent{
				mcpToolCallResultContent{
					Type: "text",
					Text: string(j),
				},
			}
			res.StructuredContent = j
		}
	} else {
		res.Content = []mcpToolCallResultContent{
			mcpToolCallResultContent{
				Type: "text",
				Text: result.Error,
			},
		}
		res.IsError = true
	}

	return jsonRpcResult[mcpToolCallResult]{
		jsonRpc: input.jsonRpc,
		Result:  res,
	}
}

func (tt *MCPTask) Attach(ctx *TaskAttachContext) {
	name_clean := strings.ToLower(strings.ReplaceAll(tt.Name, " ", "-"))
	ctx.HTTP("/mcp/"+name_clean, func(w http.ResponseWriter, r *http.Request) {
		auth, _ := url.QueryUnescape(r.URL.Query().Get("auth"))
		container := ctx.TaskContext(r.Context())
		if !ctx.VerifyAuthentication(auth) {
			w.WriteHeader(401)
			w.Write([]byte{})
			container.Logger.Log("Recieved invalid authentication over MCP.")
			return
		}

		var input jsonRpcInput
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			container.Logger.Log("Could not decode payload: %s", err.Error())
		}

		var out any

		switch input.Method {
		case "initialize":
			out = tt.initialize(&input)
		case "tools/list":
			out = tt.listTools(&input)
		case "tools/call":
			out = tt.callTool(&input, container)
		default:
			container.Logger.Log("Unknown method: %s", input.Method)
		}

		encode, _ := json.Marshal(out)
		w.Header().Add("Content-Type", "application/json")
		w.Write(encode)
	})
}

func NewMCPTask(name, version string, tools ...mcpTool) Task {
	toolDef := map[string]mcpTool{}
	for t := range tools {
		toolDef[tools[t].Name] = tools[t]
	}
	return Task{
		Name: name,
		Implementation: &MCPTask{
			Name:    name,
			Version: version,
			Tools:   toolDef,
		},
	}
}

type MCPToolType int

const (
	MCPToolTypeDestructive MCPToolType = iota
	MCPToolTypeReadOnly
	MCPToolTypeIdempotent
)

type MCPToolConfig struct {
	Title       string
	Description string
	ToolType    MCPToolType
}

func NewMCPTool[T any](fn func(input T, t *TaskContext) *MCPToolResult, cfg MCPToolConfig) mcpTool {
	name_ugly := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	last_idx := strings.LastIndex(name_ugly, ".")
	name := name_ugly[last_idx+1:]

	schema := mcpInputSchema{
		Type:       "object",
		Properties: map[string]mcpInputSchemaField{},
		Required:   []string{},
	}

	var tI T
	t := reflect.TypeOf(tI)
	for i := range t.NumField() {
		isOpt := false
		resolvedType := ""

		field := t.Field(i)
		typ := field.Type.Kind()
		if typ == reflect.Ptr {
			isOpt = true
			typ = field.Type.Elem().Kind()
		}
		switch typ {
		case reflect.String:
			resolvedType = "string"
		case reflect.Int:
			resolvedType = "number"
		default:
			continue
		}

		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if len(name) == 0 {
			name = field.Name
		}

		desc := field.Tag.Get("description")

		schema.Properties[name] = mcpInputSchemaField{
			Type:        resolvedType,
			Description: desc,
		}
		if !isOpt {
			schema.Required = append(schema.Required, name)
		}
	}

	enc, _ := json.Marshal(schema)

	title := name
	if len(cfg.Title) > 0 {
		title = cfg.Title
	}

	return mcpTool{
		Name:        name,
		InputSchema: enc,
		Description: cfg.Description,
		Annotation: mcpToolAnnotation{
			Title:           title,
			ReadOnlyHint:    cfg.ToolType == MCPToolTypeReadOnly,
			DestructiveHint: cfg.ToolType == MCPToolTypeDestructive,
			IdempotentHint:  cfg.ToolType == MCPToolTypeIdempotent,
		},
		Tool: func(input json.RawMessage, t *TaskContext) *MCPToolResult {
			var ip T
			err := json.Unmarshal(input, &ip)
			if err != nil {
				return &MCPToolResult{
					Success: false,
					Error:   "Invalid JSON sent.",
				}
			}
			return fn(ip, t)
		},
	}
}
