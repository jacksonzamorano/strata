package strata

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
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
type mcpInitialize struct {
	ProtocolVersion string          `json:"protocolVersion"`
	ServerInfo      mcpServerInfo   `json:"serverInfo"`
	Capabilities    mcpCapabilities `json:"capabilities"`
	Instructions    string          `json:"instructions,omitempty"`
}
type mcpServerInfo struct {
	Name    string    `json:"name"`
	Version string    `json:"version"`
	Icons   []mcpIcon `json:"icons,omitempty"`
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
	Format      string `json:"format,omitempty"`
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

type MCPDate struct{ time.Time }

func (d *MCPDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

type mcpToolFn = func(input json.RawMessage, ctx *TaskContext) *MCPToolResult
type mcpTool struct {
	Name        string
	Description string
	Tool        mcpToolFn
	InputSchema json.RawMessage
	Annotation  mcpToolAnnotation
}
type mcpIcon struct {
	Source   string `json:"src"`
	MimeType string `json:"mimeType"`
}
type mcpTaskIcon struct {
	Bytes  []byte
	Format string
}

type MCPTask struct {
	Name         string
	Version      string
	Instructions string
	Icon         *mcpTaskIcon
	Tools        map[string]mcpTool
}

func (tt *MCPTask) initialize(input *jsonRpcInput) jsonRpcResult[mcpInitialize] {
	icons := []mcpIcon{}
	if tt.Icon != nil {
		encoded := base64.StdEncoding.EncodeToString(tt.Icon.Bytes)
		icons = append(icons, mcpIcon{
			Source: fmt.Sprintf("data:%s;base64,%s", tt.Icon.Format, encoded),
		})
	}
	return jsonRpcResult[mcpInitialize]{
		jsonRpc: input.jsonRpc,
		Result: mcpInitialize{
			ProtocolVersion: "2025-11-25",
			ServerInfo: mcpServerInfo{
				Name:    tt.Name,
				Version: tt.Version,
				Icons:   icons,
			},
			Capabilities: mcpCapabilities{
				Tools: map[string]mcpToolDef{},
			},
			Instructions: tt.Instructions,
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
				{
					Type: "text",
					Text: str,
				},
			}
		} else {
			j, _ := json.Marshal(result.Response)
			res.Content = []mcpToolCallResultContent{
				{
					Type: "text",
					Text: string(j),
				},
			}
			res.StructuredContent = j
		}
	} else {
		res.Content = []mcpToolCallResultContent{
			{
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
		case "ping":
			out = jsonRpcResult[map[string]any]{
				jsonRpc: input.jsonRpc,
				Result:  map[string]any{},
			}
		case "notifications/initialized":
			return
		default:
			container.Logger.Log("Unknown method: %s", input.Method)
			return
		}

		encode, _ := json.Marshal(out)
		w.Header().Set("Content-Type", "application/json")
		w.Write(encode)
	})
}

type mcpModTask = func(tk *MCPTask)

func NewMCPTask(name, version string, tools ...mcpModTask) Task {
	tk := &MCPTask{
		Name:    name,
		Version: version,
		Tools:   map[string]mcpTool{},
	}
	for t := range tools {
		tools[t](tk)
	}
	return Task{
		Name:           name,
		Implementation: tk,
	}
}

func ToolSuccess(res any) *MCPToolResult {
	return &MCPToolResult{
		Response: res,
		Success:  true,
	}
}

func ToolError(msg string) *MCPToolResult {
	return &MCPToolResult{
		Error:   msg,
		Success: false,
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

func NewMCPTool[T any](fn func(input T, t *TaskContext) *MCPToolResult, cfg MCPToolConfig) mcpModTask {
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
	for field := range t.Fields() {
		isOpt := false
		resolvedType := ""
		resolvedFormat := ""

		fieldTyp := field.Type
		fieldKind := fieldTyp.Kind()
		if fieldKind == reflect.Pointer {
			isOpt = true
			fieldTyp = fieldTyp.Elem()
			fieldKind = fieldTyp.Kind()
		}
		if fieldTyp == reflect.TypeFor[MCPDate]() {
			resolvedType = "string"
			resolvedFormat = "date"
		} else {
			switch fieldKind {
			case reflect.String:
				resolvedType = "string"
			case reflect.Int:
				resolvedType = "number"
			default:
				continue
			}
		}

		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if len(name) == 0 {
			name = field.Name
		}

		desc := field.Tag.Get("description")

		schema.Properties[name] = mcpInputSchemaField{
			Type:        resolvedType,
			Format:      resolvedFormat,
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

	tool := mcpTool{
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
	return func(t *MCPTask) {
		t.Tools[tool.Name] = tool
	}
}

func MCPInstructions(ins string) mcpModTask {
	return func(m *MCPTask) {
		m.Instructions = ins
	}
}

func MCPIcon(filename string) mcpModTask {
	return func(m *MCPTask) {
		extIdx := strings.LastIndex(filename, ".")
		ext := filename[extIdx+1:]

		contents, e := os.ReadFile(filename)
		if e != nil {
			return
		}
		switch ext {
		case "png":
			m.Icon = &mcpTaskIcon{
				Bytes:  contents,
				Format: "image/png",
			}
		case "jpeg":
			m.Icon = &mcpTaskIcon{
				Bytes:  contents,
				Format: "image/jpeg",
			}
		}
	}
}
