package handlerintegrations

import (
	"encoding/json"

	"github.com/secamc93/probability/back/central/services/integrations/core/internal/domain"
)

const whatsAppIntegrationTypeID uint = 2

var protectedConfigFields = map[uint][]string{
	whatsAppIntegrationTypeID: {"phone_number_id", "waba_id", "use_platform_token"},
}

func protectConfigFields(integrationTypeID uint, incoming *map[string]any, existing *domain.Integration) {
	fields, guarded := protectedConfigFields[integrationTypeID]
	if !guarded || incoming == nil || *incoming == nil {
		return
	}

	current := map[string]any{}
	if existing != nil && len(existing.Config) > 0 {
		_ = json.Unmarshal(existing.Config, &current)
	}

	config := *incoming
	for _, field := range fields {
		delete(config, field)
		if value, ok := current[field]; ok {
			config[field] = value
		}
	}
}
