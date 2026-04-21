package args

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rafaelsoares/alfredo/internal/timeutil"
)

func RequireTwoStrings(values map[string]any, firstKey, secondKey string) (string, string, error) {
	first, err := RequireString(values, firstKey)
	if err != nil {
		return "", "", err
	}
	second, err := RequireString(values, secondKey)
	if err != nil {
		return "", "", err
	}
	return first, second, nil
}

func RequireString(values map[string]any, key string) (string, error) {
	value, ok := values[key]
	if !ok || value == nil {
		return "", fmt.Errorf("%s is required", key)
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", key)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return text, nil
}

func OptionalString(values map[string]any, key string) *string {
	value, ok := values[key]
	if !ok || value == nil {
		return nil
	}
	text, ok := value.(string)
	if !ok {
		text = fmt.Sprint(value)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	return &text
}

func RequireFloat(values map[string]any, key string) (float64, error) {
	value, ok := values[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("%s is required", key)
	}
	switch n := value.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case json.Number:
		out, err := n.Float64()
		if err != nil {
			return 0, fmt.Errorf("%s must be a number: %w", key, err)
		}
		return out, nil
	case string:
		out, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return 0, fmt.Errorf("%s must be a number: %w", key, err)
		}
		return out, nil
	default:
		return 0, fmt.Errorf("%s must be a number", key)
	}
}

func RequireInt(values map[string]any, key string) (int, error) {
	value, ok := values[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("%s is required", key)
	}
	return NumberToInt(value, key)
}

func NumberToInt(value any, key string) (int, error) {
	switch n := value.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case json.Number:
		out, err := n.Int64()
		if err != nil {
			return 0, fmt.Errorf("%s must be an integer: %w", key, err)
		}
		return int(out), nil
	case string:
		out, err := strconv.Atoi(strings.TrimSpace(n))
		if err != nil {
			return 0, fmt.Errorf("%s must be an integer: %w", key, err)
		}
		return out, nil
	default:
		return 0, fmt.Errorf("%s must be an integer", key)
	}
}

func RequireDate(values map[string]any, key string) (time.Time, error) {
	text, err := RequireString(values, key)
	if err != nil {
		return time.Time{}, err
	}
	return ParseDate(text, key)
}

func OptionalDate(values map[string]any, key string) (time.Time, error) {
	value, ok := values[key]
	if !ok || value == nil {
		return time.Time{}, nil
	}
	text, ok := value.(string)
	if !ok {
		text = fmt.Sprint(value)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return time.Time{}, nil
	}
	return ParseDate(text, key)
}

func ParseDate(text, key string) (time.Time, error) {
	out, err := time.Parse("2006-01-02", text)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be a date in YYYY-MM-DD format: %w", key, err)
	}
	return out, nil
}

func RequireUserTime(values map[string]any, key string, location *time.Location) (time.Time, error) {
	text, err := RequireString(values, key)
	if err != nil {
		return time.Time{}, err
	}
	return ParseUserTime(text, key, location)
}

func ParseUserTime(text, key string, location *time.Location) (time.Time, error) {
	out, err := timeutil.ParseUserTime(text, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", key, err)
	}
	return out, nil
}
