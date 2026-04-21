package policy

import (
	"strings"
	"testing"
)

func TestBuildIncludesSections(t *testing.T) {
	out := Build()
	for _, want := range []string{
		"Você é o Alfredo",
		"PETS:",
		"SAÚDE PESSOAL:",
		"Nunca invente identificadores.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}
