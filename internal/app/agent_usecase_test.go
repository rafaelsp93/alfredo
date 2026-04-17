package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
)

type recordingAgentRouter struct {
	systemPrompt string
	tools        []agentdomain.Tool
	inputText    string
	calls        int
}

func (r *recordingAgentRouter) Execute(
	_ context.Context,
	systemPrompt string,
	tools []agentdomain.Tool,
	inputText string,
	_ func(context.Context, agentdomain.ToolCall) (agentdomain.ToolResult, error),
) (string, agentdomain.Invocation, error) {
	r.calls++
	r.systemPrompt = systemPrompt
	r.tools = append([]agentdomain.Tool(nil), tools...)
	r.inputText = inputText
	return "resposta", agentdomain.Invocation{}, nil
}

func TestAgentUseCaseHandleUsesOneShotPrompt(t *testing.T) {
	router := &recordingAgentRouter{}
	uc := NewAgentUseCase(router, nil, nil, nil, nil, nil, nil, time.UTC, zap.NewNop())

	reply, err := uc.Handle(context.Background(), "Nutella tomou banho quando?")
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if reply != "resposta" {
		t.Fatalf("reply = %q", reply)
	}
	if router.calls != 1 {
		t.Fatalf("router calls = %d", router.calls)
	}
	if router.inputText != "Nutella tomou banho quando?" {
		t.Fatalf("input text = %q", router.inputText)
	}

	required := []string{
		"interação de uma única resposta",
		"nunca faça perguntas",
		"nunca peça esclarecimentos",
		"nunca proponha próximos passos",
		"nunca tente continuar a conversa",
		"quando foi o banho",
		"quando foi a última consulta",
		"marcar banho e tosa",
		"type=grooming",
	}
	for _, want := range required {
		if !strings.Contains(router.systemPrompt, want) {
			t.Fatalf("system prompt missing %q:\n%s", want, router.systemPrompt)
		}
	}
	if strings.Contains(router.systemPrompt, "esclarecimento em vez de chamar uma ferramenta") {
		t.Fatalf("system prompt still allows clarification questions:\n%s", router.systemPrompt)
	}
}

func TestBuildAgentToolsAppointmentMetadata(t *testing.T) {
	tools := buildAgentTools()

	listAppointments := toolByName(t, tools, "list_appointments")
	for _, want := range []string{"banho", "banho e tosa", "tosa", "grooming", "quando foi a última consulta"} {
		if !strings.Contains(strings.ToLower(listAppointments.Description), strings.ToLower(want)) {
			t.Fatalf("list_appointments description missing %q: %q", want, listAppointments.Description)
		}
	}

	scheduleAppointment := toolByName(t, tools, "schedule_appointment")
	props, ok := scheduleAppointment.InputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schedule_appointment properties has unexpected type: %#v", scheduleAppointment.InputSchema["properties"])
	}
	typeProp, ok := props["type"].(map[string]any)
	if !ok {
		t.Fatalf("schedule_appointment type schema has unexpected type: %#v", props["type"])
	}
	desc, ok := typeProp["description"].(string)
	if !ok {
		t.Fatalf("schedule_appointment type description missing: %#v", typeProp)
	}
	for _, want := range []string{"banho", "banho e tosa", "tosa", "grooming"} {
		if !strings.Contains(strings.ToLower(desc), strings.ToLower(want)) {
			t.Fatalf("schedule_appointment type description missing %q: %q", want, desc)
		}
	}
}

func TestAgentPromptAndToolsCoverRepresentativeUtterances(t *testing.T) {
	tools := buildAgentTools()
	prompt := agentSystemPrompt

	tests := []struct {
		name string
		want []string
	}{
		{
			name: "bath history",
			want: []string{"quando foi o banho", "list_appointments"},
		},
		{
			name: "bath booking",
			want: []string{"marcar banho e tosa", "schedule_appointment", "type=grooming"},
		},
		{
			name: "consult history",
			want: []string{"quando foi a última consulta", "list_appointments"},
		},
		{
			name: "observation logging",
			want: []string{"observação", "log_observation"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			text := prompt + "\n" + allToolText(tools)
			for _, want := range tc.want {
				if !strings.Contains(strings.ToLower(text), strings.ToLower(want)) {
					t.Fatalf("combined prompt/tool text missing %q:\n%s", want, text)
				}
			}
		})
	}
}

func toolByName(t *testing.T, tools []agentdomain.Tool, name string) agentdomain.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool %q not found", name)
	return agentdomain.Tool{}
}

func allToolText(tools []agentdomain.Tool) string {
	var b strings.Builder
	for _, tool := range tools {
		b.WriteString(tool.Name)
		b.WriteString("\n")
		b.WriteString(tool.Description)
		b.WriteString("\n")
	}
	return b.String()
}
