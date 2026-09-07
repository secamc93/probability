package usecasetemplates

import (
	"strings"

	"github.com/secamc93/probability/back/central/services/integrations/messaging/whatsapp/internal/domain/ports"
)

var platformOnlyTemplates = map[string]bool{
	"alerta_servidor":            true,
	"recuperacion_codigo":        true,
	"reporte_saldo_billetera":    true,
	"reporte_saldo_billetera_v2": true,
	"resumen_pago_suscripcion":   true,
	"hello_world":                true,
}

func isPlatformOnlyTemplate(name string) bool {
	return platformOnlyTemplates[strings.ToLower(strings.TrimSpace(name))]
}

func componentsForCreate(template ports.TemplateDefinitionRemote) []map[string]any {
	if !strings.EqualFold(strings.TrimSpace(template.Category), "AUTHENTICATION") {
		return template.Components
	}
	return authenticationComponents(template.Components)
}

func authenticationComponents(components []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(components))

	for _, component := range components {
		switch strings.ToUpper(stringField(component, "type")) {
		case "BODY":
			body := map[string]any{"type": "BODY"}
			if recommendation, ok := component["add_security_recommendation"].(bool); ok && recommendation {
				body["add_security_recommendation"] = true
			}
			out = append(out, body)
		case "FOOTER":
			minutes := intField(component, "code_expiration_minutes")
			if minutes <= 0 {
				continue
			}
			out = append(out, map[string]any{
				"type":                    "FOOTER",
				"code_expiration_minutes": minutes,
			})
		case "BUTTONS":
			buttons := otpButtons(component)
			if len(buttons) == 0 {
				continue
			}
			out = append(out, map[string]any{"type": "BUTTONS", "buttons": buttons})
		}
	}

	return out
}

func otpButtons(component map[string]any) []map[string]any {
	raw, ok := component["buttons"].([]any)
	if !ok {
		return nil
	}

	buttons := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		button, ok := item.(map[string]any)
		if !ok {
			continue
		}

		otp := map[string]any{
			"type":     "OTP",
			"otp_type": otpTypeFrom(button),
		}
		if text := stringField(button, "text"); text != "" {
			otp["text"] = text
		}
		if autofill := stringField(button, "autofill_text"); autofill != "" {
			otp["autofill_text"] = autofill
		}
		buttons = append(buttons, otp)
	}

	return buttons
}

func otpTypeFrom(button map[string]any) string {
	if declared := strings.ToUpper(stringField(button, "otp_type")); declared != "" {
		return declared
	}

	url := stringField(button, "url")
	switch {
	case strings.Contains(url, "otp_type=ONE_TAP"):
		return "ONE_TAP"
	case strings.Contains(url, "otp_type=ZERO_TAP"):
		return "ZERO_TAP"
	default:
		return "COPY_CODE"
	}
}

func stringField(source map[string]any, key string) string {
	value, _ := source[key].(string)
	return strings.TrimSpace(value)
}

func intField(source map[string]any, key string) int {
	switch value := source[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	}
	return 0
}
