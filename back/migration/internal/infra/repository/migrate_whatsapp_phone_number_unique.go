package repository

import (
	"context"
	"fmt"
)

func (r *Repository) migrateWhatsappPhoneNumberUnique(ctx context.Context) error {
	db := r.db.Conn(ctx)

	if err := db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS uq_integrations_whatsapp_phone_number_id
    ON integrations ((config ->> 'phone_number_id'))
    WHERE integration_type_id = 2
      AND deleted_at IS NULL
      AND coalesce(config ->> 'phone_number_id', '') <> ''
`).Error; err != nil {
		return fmt.Errorf("unique index on whatsapp phone_number_id: %w", err)
	}

	return nil
}
