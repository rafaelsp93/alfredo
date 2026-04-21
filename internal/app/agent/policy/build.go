package policy

import "strings"

func Build() string {
	return strings.Join([]string{
		base,
		"",
		pets,
		"",
		health,
		"",
		"Nunca invente identificadores.",
		"Se o pedido do Rafael estiver ambíguo ou faltar informação essencial, responda apenas que não conseguiu concluir o pedido.",
	}, "\n")
}
