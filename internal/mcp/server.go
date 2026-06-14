package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/alextacho/forge/internal/discovery"
	"github.com/alextacho/forge/internal/guide"
	"github.com/alextacho/forge/internal/project"
	"github.com/alextacho/forge/internal/version"
)

type Server struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Getwd  func() (string, error)
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallParams struct {
	Name string `json:"name"`
}

type resourceReadParams struct {
	URI string `json:"uri"`
}

func Serve(server Server) int {
	if server.Stdin == nil {
		server.Stdin = os.Stdin
	}
	if server.Stdout == nil {
		server.Stdout = os.Stdout
	}
	if server.Stderr == nil {
		server.Stderr = os.Stderr
	}
	if server.Getwd == nil {
		server.Getwd = os.Getwd
	}
	if err := serve(server); err != nil {
		fmt.Fprintln(server.Stderr, "Error:", err)
		return 1
	}
	return 0
}

func serve(server Server) error {
	reader := bufio.NewReader(server.Stdin)
	for {
		payload, err := readMessage(reader)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		var req request
		if err := json.Unmarshal(payload, &req); err != nil {
			return fmt.Errorf("decode request: %w", err)
		}
		if len(req.ID) == 0 {
			continue
		}
		result, rpcErr := handle(server, req)
		resp := response{JSONRPC: "2.0", ID: req.ID, Result: result, Error: rpcErr}
		if rpcErr != nil {
			resp.Result = nil
		}
		if err := writeMessage(server.Stdout, resp); err != nil {
			return err
		}
	}
}

func handle(server Server, req request) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools":     map[string]any{"listChanged": false},
				"resources": map[string]any{"subscribe": false, "listChanged": false},
			},
			"serverInfo": map[string]any{"name": "forge", "version": version.Version},
		}, nil
	case "tools/list":
		return map[string]any{"tools": tools()}, nil
	case "tools/call":
		var params toolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return callTool(server, params.Name), nil
	case "resources/list":
		return map[string]any{"resources": resources()}, nil
	case "resources/read":
		var params resourceReadParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return readResource(server, params.URI)
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found"}
	}
}

func tools() []map[string]any {
	noArgs := map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}
	return []map[string]any{
		{
			"name":        "forge_instructions",
			"description": "Return agent-oriented Forge setup and snapshot instructions.",
			"inputSchema": noArgs,
		},
		{
			"name":        "forge_config_schema",
			"description": "Return the .forge/config.yaml shape and validation rules.",
			"inputSchema": noArgs,
		},
		{
			"name":        "forge_project_status",
			"description": "Inspect the current Forge project, configured paths, snapshots, and templates.",
			"inputSchema": noArgs,
		},
		{
			"name":        "forge_list_snapshots",
			"description": "List valid snapshots in the current Forge project.",
			"inputSchema": noArgs,
		},
	}
}

func resources() []map[string]any {
	return []map[string]any{
		{"uri": "forge://instructions", "name": "Forge instructions", "mimeType": "text/plain"},
		{"uri": "forge://config-schema", "name": "Forge config schema", "mimeType": "text/plain"},
		{"uri": "forge://status", "name": "Forge project status", "mimeType": "text/plain"},
	}
}

func callTool(server Server, name string) map[string]any {
	switch name {
	case "forge_instructions":
		return textToolResult(guide.Instructions, false)
	case "forge_config_schema":
		return textToolResult(guide.ConfigSchema, false)
	case "forge_project_status":
		text, err := projectStatusText(server)
		if err != nil {
			return textToolResult(err.Error(), true)
		}
		return textToolResult(text, false)
	case "forge_list_snapshots":
		found, err := findProject(server)
		if err != nil {
			return textToolResult(err.Error(), true)
		}
		snapshots, err := discovery.ListSnapshots(found.Root)
		if err != nil {
			return textToolResult(err.Error(), true)
		}
		payload, err := json.MarshalIndent(map[string]any{"snapshots": snapshots}, "", "  ")
		if err != nil {
			return textToolResult(err.Error(), true)
		}
		return textToolResult(string(payload), false)
	default:
		return textToolResult("unknown tool "+strconv.Quote(name), true)
	}
}

func readResource(server Server, uri string) (any, *rpcError) {
	var text string
	switch uri {
	case "forge://instructions":
		text = guide.Instructions
	case "forge://config-schema":
		text = guide.ConfigSchema
	case "forge://status":
		status, err := projectStatusText(server)
		if err != nil {
			return nil, &rpcError{Code: -32000, Message: err.Error()}
		}
		text = status
	default:
		return nil, &rpcError{Code: -32602, Message: "unknown resource URI"}
	}
	return map[string]any{
		"contents": []map[string]any{{
			"uri":      uri,
			"mimeType": "text/plain",
			"text":     text,
		}},
	}, nil
}

func textToolResult(text string, isError bool) map[string]any {
	return map[string]any{
		"content": []map[string]string{{"type": "text", "text": text}},
		"isError": isError,
	}
}

func projectStatusText(server Server) (string, error) {
	found, err := findProject(server)
	if err != nil {
		return "", err
	}
	status, err := discovery.Inspect(found)
	if err != nil {
		return "", err
	}
	return discovery.FormatStatus(status), nil
}

func findProject(server Server) (project.Project, error) {
	workingDirectory, err := server.Getwd()
	if err != nil {
		return project.Project{}, fmt.Errorf("get working directory: %w", err)
	}
	return project.Find(workingDirectory)
}

func invalidParams(err error) *rpcError {
	return &rpcError{Code: -32602, Message: "invalid params: " + err.Error()}
}

func readMessage(reader *bufio.Reader) ([]byte, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(strings.TrimSpace(line), "{") {
		return []byte(line), nil
	}

	var length int
	for {
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			break
		}
		name, value, ok := strings.Cut(trimmed, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %w", err)
			}
			length = parsed
		}
		line, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
	}
	if length <= 0 {
		return nil, errors.New("missing Content-Length header")
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeMessage(writer io.Writer, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	var message bytes.Buffer
	fmt.Fprintf(&message, "Content-Length: %d\r\n\r\n", len(payload))
	message.Write(payload)
	_, err = writer.Write(message.Bytes())
	return err
}
