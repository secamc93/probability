package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/secamc93/probability/back/central/services/integrations/ecommerce/woocommerce/internal/domain"
)

func (uc *wooCommerceUseCase) UpdateInventory(ctx context.Context, integrationID string, productExternalID string, quantity int) error {
	integration, err := uc.service.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		return fmt.Errorf("getting integration: %w", err)
	}
	if integration == nil {
		return domain.ErrIntegrationNotFound
	}

	if enabled, _ := integration.Config["inventory_sync_enabled"].(bool); !enabled {
		uc.logger.Info(ctx).
			Str("integration_id", integrationID).
			Msg("Sync de inventario desactivado para la integracion WooCommerce, push omitido")
		return nil
	}

	storeURL, err := extractString(integration.Config, "store_url")
	if err != nil {
		return domain.ErrMissingStoreURL
	}
	storeURL = resolveEffectiveStoreURL(integration, storeURL)

	consumerKey, err := uc.service.DecryptCredential(ctx, integrationID, "consumer_key")
	if err != nil {
		return fmt.Errorf("decrypting consumer_key: %w", err)
	}
	consumerSecret, err := uc.service.DecryptCredential(ctx, integrationID, "consumer_secret")
	if err != nil {
		return fmt.Errorf("decrypting consumer_secret: %w", err)
	}

	if err := uc.client.UpdateProductStock(ctx, storeURL, consumerKey, consumerSecret, productExternalID, quantity); err != nil {
		if errors.Is(err, domain.ErrProductNotFoundInStore) && integration.BusinessID != nil {
			if healErr := uc.healStaleMapping(ctx, integration, storeURL, consumerKey, consumerSecret, productExternalID, quantity); healErr == nil {
				return nil
			}
		}
		uc.logger.Error(ctx).
			Err(err).
			Str("integration_id", integrationID).
			Str("external_product_id", productExternalID).
			Int("quantity", quantity).
			Msg("Error al actualizar stock en WooCommerce")
		return err
	}

	uc.logger.Info(ctx).
		Str("integration_id", integrationID).
		Str("external_product_id", productExternalID).
		Int("quantity", quantity).
		Msg("Stock actualizado en WooCommerce")

	return nil
}

func (uc *wooCommerceUseCase) healStaleMapping(ctx context.Context, integration *domain.Integration, storeURL, consumerKey, consumerSecret, productExternalID string, quantity int) error {
	parent, variant, _ := strings.Cut(productExternalID, ":")

	probabilityProductID, err := uc.productRepo.GetProductIDByExternalRef(ctx, integration.ID, parent, variant)
	if err != nil || probabilityProductID == "" {
		return fmt.Errorf("mapeo obsoleto sin producto asociado en Probability")
	}

	sku, err := uc.productRepo.GetProductSKUByID(ctx, probabilityProductID, *integration.BusinessID)
	if err != nil || sku == "" {
		return fmt.Errorf("no se pudo obtener el SKU del producto para autocorregir el mapeo")
	}

	correlationID := uuid.New().String()
	integrationIDStr := fmt.Sprintf("%d", integration.ID)
	if err := uc.AssociateProducts(ctx, integrationIDStr, *integration.BusinessID, correlationID, []string{sku}); err != nil {
		return fmt.Errorf("re-asociando producto por SKU: %w", err)
	}

	newParent, newVariant, err := uc.productRepo.GetExternalRefs(ctx, probabilityProductID, integration.ID)
	if err != nil || newParent == "" {
		return fmt.Errorf("la re-asociacion no encontro un ID nuevo en la tienda")
	}
	newExternalID := newParent
	if newVariant != "" {
		newExternalID = newParent + ":" + newVariant
	}
	if newExternalID == productExternalID {
		return fmt.Errorf("el ID sigue siendo el mismo tras re-asociar")
	}

	if err := uc.client.UpdateProductStock(ctx, storeURL, consumerKey, consumerSecret, newExternalID, quantity); err != nil {
		return fmt.Errorf("reintentando con el ID corregido: %w", err)
	}

	uc.logger.Info(ctx).
		Str("sku", sku).
		Str("external_product_id_anterior", productExternalID).
		Str("external_product_id_nuevo", newExternalID).
		Int("quantity", quantity).
		Msg("Mapeo de WooCommerce obsoleto, autocorregido y stock actualizado")

	return nil
}
