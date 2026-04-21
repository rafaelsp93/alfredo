package registry

import (
	"context"
	"errors"
	"strings"
	"testing"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
)

type stubHandler struct {
	spec agentdomain.Tool
	err  error
	out  any
}

func (h stubHandler) Spec() agentdomain.Tool { return h.spec }

func (h stubHandler) Handle(context.Context, map[string]any) (any, error) {
	return h.out, h.err
}

func TestNewRejectsDuplicateNames(t *testing.T) {
	_, err := New(
		stubHandler{spec: Tool("dup", "one", ObjectSchema(nil, nil))},
		stubHandler{spec: Tool("dup", "two", ObjectSchema(nil, nil))},
	)
	if err == nil || !strings.Contains(err.Error(), `duplicate tool "dup"`) {
		t.Fatalf("err = %v, want duplicate tool error", err)
	}
}

func TestExecuteUnknownTool(t *testing.T) {
	reg := MustNew(stubHandler{spec: Tool("known", "desc", ObjectSchema(nil, nil))})
	result, err := reg.Execute(context.Background(), agentdomain.ToolCall{ID: "call-1", Name: "missing"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !result.IsError || !strings.Contains(result.Content, `unknown tool "missing"`) {
		t.Fatalf("result = %#v", result)
	}
}

func TestToolsOrderIsStable(t *testing.T) {
	reg := MustNew(
		stubHandler{spec: Tool("first", "desc", ObjectSchema(nil, nil))},
		stubHandler{spec: Tool("second", "desc", ObjectSchema(nil, nil))},
	)
	tools := reg.Tools()
	if len(tools) != 2 || tools[0].Name != "first" || tools[1].Name != "second" {
		t.Fatalf("tools = %#v", tools)
	}
}

func TestMustNewPanicsAndExecutePropagatesErrors(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = MustNew(
		stubHandler{spec: Tool("dup", "one", ObjectSchema(nil, nil))},
		stubHandler{spec: Tool("dup", "two", ObjectSchema(nil, nil))},
	)
}

func TestExecuteHandlerAndMarshalErrors(t *testing.T) {
	reg := MustNew(stubHandler{spec: Tool("boom", "desc", ObjectSchema(nil, nil)), err: errors.New("boom")})
	if _, err := reg.Execute(context.Background(), agentdomain.ToolCall{ID: "1", Name: "boom"}); err == nil {
		t.Fatal("expected handler error")
	}
	reg = MustNew(stubHandler{spec: Tool("marshal", "desc", ObjectSchema(nil, nil)), out: make(chan int)})
	if _, err := reg.Execute(context.Background(), agentdomain.ToolCall{ID: "1", Name: "marshal"}); err == nil {
		t.Fatal("expected marshal error")
	}
}
