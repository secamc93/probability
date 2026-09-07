package handlerintegrations

import (
	"testing"

	"github.com/secamc93/probability/back/central/services/integrations/core/internal/domain"
	"gorm.io/datatypes"
)

func TestProtectConfigFieldsIgnoraElNumeroEnviadoPorElCliente(t *testing.T) {
	existing := &domain.Integration{
		IntegrationTypeID: whatsAppIntegrationTypeID,
		Config:            datatypes.JSON([]byte(`{"use_platform_token":true,"test_phone_number":"3001112233"}`)),
	}

	incoming := map[string]any{
		"phone_number_id":    "1077369948787698",
		"waba_id":            "1302830408357767",
		"use_platform_token": false,
		"test_phone_number":  "3009998877",
	}

	protectConfigFields(whatsAppIntegrationTypeID, &incoming, existing)

	if _, ok := incoming["phone_number_id"]; ok {
		t.Fatal("el phone_number_id enviado por el cliente no debe llegar al config")
	}
	if _, ok := incoming["waba_id"]; ok {
		t.Fatal("el waba_id enviado por el cliente no debe llegar al config")
	}
	if incoming["use_platform_token"] != true {
		t.Fatalf("use_platform_token debe conservar el valor guardado: %v", incoming["use_platform_token"])
	}
	if incoming["test_phone_number"] != "3009998877" {
		t.Fatalf("los campos no protegidos sí se actualizan: %v", incoming["test_phone_number"])
	}
}

func TestProtectConfigFieldsNoTocaOtrosTipos(t *testing.T) {
	incoming := map[string]any{"phone_number_id": "123", "store_id": "abc"}

	protectConfigFields(1, &incoming, nil)

	if incoming["phone_number_id"] != "123" || incoming["store_id"] != "abc" {
		t.Fatalf("solo WhatsApp tiene campos protegidos: %v", incoming)
	}
}
