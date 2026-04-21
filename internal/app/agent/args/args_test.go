package args

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestArgsHelpers(t *testing.T) {
	location := time.FixedZone("BRT", -3*60*60)
	values := map[string]any{
		"name":        " Nutella ",
		"float64":     1.5,
		"floatString": "2.5",
		"int64":       float64(3),
		"intString":   "4",
		"jsonNumber":  json.Number("5"),
		"date":        "2026-04-21",
		"datetime":    "2026-04-21T10:30:00",
		"blank":       " ",
		"first":       "a",
		"second":      "b",
	}

	if first, second, err := RequireTwoStrings(values, "first", "second"); err != nil || first != "a" || second != "b" {
		t.Fatalf("RequireTwoStrings = %q %q %v", first, second, err)
	}
	if got, err := RequireString(values, "name"); err != nil || got != "Nutella" {
		t.Fatalf("RequireString = %q %v", got, err)
	}
	if got := OptionalString(values, "name"); got == nil || *got != "Nutella" {
		t.Fatalf("OptionalString name = %#v", got)
	}
	if got := OptionalString(values, "blank"); got != nil {
		t.Fatalf("OptionalString blank = %#v", got)
	}
	if got, err := RequireFloat(values, "float64"); err != nil || got != 1.5 {
		t.Fatalf("RequireFloat float64 = %v %v", got, err)
	}
	if got, err := RequireFloat(values, "floatString"); err != nil || got != 2.5 {
		t.Fatalf("RequireFloat string = %v %v", got, err)
	}
	if got, err := RequireFloat(values, "jsonNumber"); err != nil || got != 5 {
		t.Fatalf("RequireFloat jsonNumber = %v %v", got, err)
	}
	if got, err := RequireInt(values, "int64"); err != nil || got != 3 {
		t.Fatalf("RequireInt = %v %v", got, err)
	}
	if got, err := NumberToInt(values["intString"], "intString"); err != nil || got != 4 {
		t.Fatalf("NumberToInt string = %v %v", got, err)
	}
	if got, err := NumberToInt(values["jsonNumber"], "jsonNumber"); err != nil || got != 5 {
		t.Fatalf("NumberToInt jsonNumber = %v %v", got, err)
	}
	if got, err := RequireDate(values, "date"); err != nil || got.Format("2006-01-02") != "2026-04-21" {
		t.Fatalf("RequireDate = %v %v", got, err)
	}
	if got, err := OptionalDate(values, "date"); err != nil || got.Format("2006-01-02") != "2026-04-21" {
		t.Fatalf("OptionalDate = %v %v", got, err)
	}
	if got, err := OptionalDate(values, "missing"); err != nil || !got.IsZero() {
		t.Fatalf("OptionalDate missing = %v %v", got, err)
	}
	if got, err := ParseDate("2026-04-22", "date"); err != nil || got.Format("2006-01-02") != "2026-04-22" {
		t.Fatalf("ParseDate = %v %v", got, err)
	}
	if got, err := RequireUserTime(values, "datetime", location); err != nil || got.Location() != location {
		t.Fatalf("RequireUserTime = %v %v", got, err)
	}
	if got, err := ParseUserTime("2026-04-21T11:00:00", "datetime", location); err != nil || got.Location() != location {
		t.Fatalf("ParseUserTime = %v %v", got, err)
	}
}

func TestArgsHelpersErrors(t *testing.T) {
	if _, _, err := RequireTwoStrings(map[string]any{"first": "a"}, "first", "second"); err == nil {
		t.Fatal("expected second string error")
	}
	if _, err := RequireString(map[string]any{}, "name"); err == nil {
		t.Fatal("expected missing string error")
	}
	if _, err := RequireString(map[string]any{"name": 1}, "name"); err == nil {
		t.Fatal("expected type error")
	}
	if _, err := RequireFloat(map[string]any{"value": true}, "value"); err == nil {
		t.Fatal("expected float error")
	}
	if _, err := RequireFloat(map[string]any{}, "value"); err == nil {
		t.Fatal("expected missing float error")
	}
	if _, err := RequireInt(map[string]any{"value": true}, "value"); err == nil {
		t.Fatal("expected int error")
	}
	if _, err := RequireInt(map[string]any{}, "value"); err == nil {
		t.Fatal("expected missing int error")
	}
	if _, err := NumberToInt(true, "value"); err == nil {
		t.Fatal("expected NumberToInt type error")
	}
	if _, err := RequireDate(map[string]any{}, "date"); err == nil {
		t.Fatal("expected missing date error")
	}
	if _, err := ParseDate("bad", "date"); err == nil || !strings.Contains(err.Error(), "YYYY-MM-DD") {
		t.Fatalf("ParseDate err = %v", err)
	}
	if _, err := OptionalDate(map[string]any{"date": "bad"}, "date"); err == nil {
		t.Fatal("expected OptionalDate parse error")
	}
	if _, err := RequireUserTime(map[string]any{}, "datetime", time.UTC); err == nil {
		t.Fatal("expected missing datetime error")
	}
	if _, err := ParseUserTime("bad", "datetime", time.UTC); err == nil || !strings.Contains(err.Error(), "datetime") {
		t.Fatalf("ParseUserTime err = %v", err)
	}
}
