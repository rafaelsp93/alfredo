package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/telegram"
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
	uc := NewAgentUseCase(router, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, zap.NewNop())

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
		"get_pet_summary",
		"send_telegram",
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

func TestBuildAgentToolsDailyDigestMetadata(t *testing.T) {
	tools := buildAgentTools()

	summary := toolByName(t, tools, "get_pet_summary")
	required, ok := summary.InputSchema["required"].([]string)
	if !ok {
		t.Fatalf("get_pet_summary required has unexpected type: %#v", summary.InputSchema["required"])
	}
	if len(required) != 0 {
		t.Fatalf("get_pet_summary should not require input: %#v", summary.InputSchema)
	}
	for _, want := range []string{"all-pets", "daily digest", "resumo diário", "supplies needing reorder"} {
		if !strings.Contains(strings.ToLower(summary.Description), strings.ToLower(want)) {
			t.Fatalf("get_pet_summary description missing %q: %q", want, summary.Description)
		}
	}

	send := toolByName(t, tools, "send_telegram")
	sendRequired, ok := send.InputSchema["required"].([]string)
	if !ok || !sameStringSet(sendRequired, []string{"message"}) {
		t.Fatalf("send_telegram required = %#v", send.InputSchema["required"])
	}
}

func TestAgentUseCaseSendTelegramIsBestEffort(t *testing.T) {
	uc := NewAgentUseCase(nil, nil, nil, nil, nil, nil, nil, nil, failingTelegram{err: errors.New("telegram down")}, time.UTC, zap.NewNop())

	result, err := uc.DispatchToolCall(context.Background(), agentdomain.ToolCall{
		ID:        "call-1",
		Name:      "send_telegram",
		Arguments: map[string]any{"message": "Resumo dos pets"},
	})
	if err != nil {
		t.Fatalf("DispatchToolCall returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("send_telegram returned error tool result: %#v", result)
	}
	if !strings.Contains(result.Content, "Não consegui enviar") {
		t.Fatalf("send_telegram result content = %q", result.Content)
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

type failingTelegram struct {
	err error
}

func (t failingTelegram) Send(context.Context, telegram.Message) error {
	return t.err
}

func sameStringSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[string]int, len(got))
	for _, value := range got {
		seen[value]++
	}
	for _, value := range want {
		seen[value]--
		if seen[value] < 0 {
			return false
		}
	}
	return true
}
