package repository

import (
	"context"
	"fmt"
)

func (r *Repository) migrateUserGoogleID(ctx context.Context) error {
	db := r.db.Conn(ctx)

	if err := db.Exec(`
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS google_id VARCHAR(64)
`).Error; err != nil {
		return fmt.Errorf("add user.google_id: %w", err)
	}

	if err := db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_google_id
    ON "user" (google_id)
    WHERE google_id IS NOT NULL AND deleted_at IS NULL
`).Error; err != nil {
		return fmt.Errorf("unique index on user.google_id: %w", err)
	}

	return nil
}
