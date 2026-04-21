package registry

import (
	"context"
	"encoding/json"
	"fmt"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
)

type ToolHandler interface {
	Spec() agentdomain.Tool
	Handle(ctx context.Context, args map[string]any) (any, error)
}

type ToolRegistry interface {
	Tools() []agentdomain.Tool
	Execute(ctx context.Context, call agentdomain.ToolCall) (agentdomain.ToolResult, error)
}

type toolRegistry struct {
	tools    []agentdomain.Tool
	handlers map[string]ToolHandler
}

func New(handlers ...ToolHandler) (ToolRegistry, error) {
	tools := make([]agentdomain.Tool, 0, len(handlers))
	byName := make(map[string]ToolHandler, len(handlers))
	for _, handler := range handlers {
		spec := handler.Spec()
		if spec.Name == "" {
			return nil, fmt.Errorf("tool name is required")
		}
		if _, exists := byName[spec.Name]; exists {
			return nil, fmt.Errorf("duplicate tool %q", spec.Name)
		}
		tools = append(tools, spec)
		byName[spec.Name] = handler
	}
	return &toolRegistry{tools: tools, handlers: byName}, nil
}

func MustNew(handlers ...ToolHandler) ToolRegistry {
	reg, err := New(handlers...)
	if err != nil {
		panic(err)
	}
	return reg
}

func (r *toolRegistry) Tools() []agentdomain.Tool {
	return append([]agentdomain.Tool(nil), r.tools...)
}

func (r *toolRegistry) Execute(ctx context.Context, call agentdomain.ToolCall) (agentdomain.ToolResult, error) {
	handler, ok := r.handlers[call.Name]
	if !ok {
		err := fmt.Errorf("unknown tool %q", call.Name)
		return ErrorToolResult(call, err), err
	}
	result, err := handler.Handle(ctx, call.Arguments)
	if err != nil {
		return ErrorToolResult(call, err), err
	}
	content, err := json.Marshal(result)
	if err != nil {
		marshalErr := fmt.Errorf("marshal tool result for %q: %w", call.Name, err)
		return ErrorToolResult(call, marshalErr), marshalErr
	}
	return agentdomain.ToolResult{CallID: call.ID, Content: string(content)}, nil
}

func ErrorToolResult(call agentdomain.ToolCall, err error) agentdomain.ToolResult {
	return agentdomain.ToolResult{CallID: call.ID, Content: err.Error(), IsError: true}
}

func Tool(name, description string, schema map[string]any) agentdomain.Tool {
	return agentdomain.Tool{Name: name, Description: description, InputSchema: schema}
}

func ObjectSchema(props map[string]any, required []string) map[string]any {
	if props == nil {
		props = map[string]any{}
	}
	return map[string]any{"type": "object", "properties": props, "required": required}
}

func Properties(kv ...string) map[string]any {
	props := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		props[kv[i]] = map[string]any{"type": kv[i+1]}
	}
	return props
}
