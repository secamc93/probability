package queue

import (
	"context"
	"testing"

	"github.com/secamc93/probability/back/central/services/integrations/ecommerce/woocommerce/internal/app/usecases"
	"github.com/secamc93/probability/back/central/shared/log"
)

type fakeUseCase struct {
	usecases.IWooCommerceUseCase
	gotIntegrationID string
	gotExternalID    string
	gotQuantity      int
}

func (f *fakeUseCase) UpdateInventory(ctx context.Context, integrationID string, productExternalID string, quantity int) error {
	f.gotIntegrationID = integrationID
	f.gotExternalID = productExternalID
	f.gotQuantity = quantity
	return nil
}

func TestInventoryPushConsumer_HandleComposesVariationRef(t *testing.T) {
	fake := &fakeUseCase{}
	c := NewInventoryPushConsumer(nil, fake, log.New())

	body := []byte(`{"product_id":"p1","external_product_id":"123","external_variant_id":"456","integration_id":7,"quantity":10}`)
	c.handle(context.Background(), body)

	if fake.gotExternalID != "123:456" {
		t.Errorf("external_product_id enviado al usecase = %q, want %q", fake.gotExternalID, "123:456")
	}
	if fake.gotIntegrationID != "7" {
		t.Errorf("integration_id = %q, want %q", fake.gotIntegrationID, "7")
	}
	if fake.gotQuantity != 10 {
		t.Errorf("quantity = %d, want 10", fake.gotQuantity)
	}
}

func TestInventoryPushConsumer_HandleWithoutVariant(t *testing.T) {
	fake := &fakeUseCase{}
	c := NewInventoryPushConsumer(nil, fake, log.New())

	body := []byte(`{"product_id":"p1","external_product_id":"123","integration_id":7,"quantity":5}`)
	c.handle(context.Background(), body)

	if fake.gotExternalID != "123" {
		t.Errorf("external_product_id enviado al usecase = %q, want %q", fake.gotExternalID, "123")
	}
}
