package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServeToolsListAndCall(t *testing.T) {
	root := newProject(t)
	input := framed(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
		framed(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`) +
		framed(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"forge_project_status","arguments":{}}}`)

	var stdout, stderr bytes.Buffer
	code := Serve(Server{
		Stdin:  strings.NewReader(input),
		Stdout: &stdout,
		Stderr: &stderr,
		Getwd:  func() (string, error) { return root, nil },
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}

	responses := decodeResponses(t, stdout.String())
	if len(responses) != 3 {
		t.Fatalf("responses=%d", len(responses))
	}
	if !strings.Contains(string(responses[1]), "forge_instructions") {
		t.Fatalf("tools/list response missing tool: %s", responses[1])
	}
	if !strings.Contains(string(responses[2]), "Managed paths: 1") {
		t.Fatalf("status response missing project state: %s", responses[2])
	}
}

func TestServeResourcesRead(t *testing.T) {
	input := framed(`{"jsonrpc":"2.0","id":"read","method":"resources/read","params":{"uri":"forge://instructions"}}`)
	var stdout, stderr bytes.Buffer
	code := Serve(Server{
		Stdin:  strings.NewReader(input),
		Stdout: &stdout,
		Stderr: &stderr,
		Getwd:  func() (string, error) { return t.TempDir(), nil },
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	responses := decodeResponses(t, stdout.String())
	if len(responses) != 1 {
		t.Fatalf("responses=%d", len(responses))
	}
	if !strings.Contains(string(responses[0]), "Snapshot workflow") {
		t.Fatalf("resource response missing instructions: %s", responses[0])
	}
}

func framed(payload string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(payload), payload)
}

func decodeResponses(t *testing.T, output string) []json.RawMessage {
	t.Helper()
	reader := strings.NewReader(output)
	var responses []json.RawMessage
	for reader.Len() > 0 {
		header, err := readHeaderLine(reader)
		if err != nil {
			t.Fatal(err)
		}
		name, value, ok := strings.Cut(strings.TrimSpace(header), ":")
		if !ok || !strings.EqualFold(name, "Content-Length") {
			t.Fatalf("bad header %q", header)
		}
		var length int
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &length); err != nil {
			t.Fatal(err)
		}
		blank, err := readHeaderLine(reader)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(blank) != "" {
			t.Fatalf("missing blank header line: %q", blank)
		}
		payload := make([]byte, length)
		if _, err := io.ReadFull(reader, payload); err != nil {
			t.Fatal(err)
		}
		responses = append(responses, append(json.RawMessage(nil), payload...))
	}
	return responses
}

func readHeaderLine(reader *strings.Reader) (string, error) {
	var line strings.Builder
	for {
		character, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}
		line.WriteRune(character)
		if character == '\n' {
			return line.String(), nil
		}
	}
}

func newProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".forge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "workspace"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := []byte("paths:\n  - workspace\n")
	if err := os.WriteFile(filepath.Join(root, ".forge", "config.yaml"), config, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
