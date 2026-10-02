package store

import (
	"errors"
	"time"

	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

// Save stores (or overwrites) the OTP for an email, valid for 5 minutes.
func Save(database db.Database, email, code string) error {
	expires := time.Now().Add(5 * time.Minute)
	return database.Exec(`
		INSERT INTO otps (email, code, expires_at)
		VALUES (?, ?, ?)
		ON CONFLICT (email) DO UPDATE SET code = ?, expires_at = ?
	`, email, code, expires, code, expires)
}

// Verify checks the OTP for an email and deletes it on success (one-time use).
func Verify(database db.Database, email, code string) (bool, error) {
	var otp models.Otp
	err := database.Where("email = ?", email).First(&otp)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if time.Now().After(otp.ExpiresAt) {
		return false, nil
	}
	if otp.Code != code {
		return false, nil
	}

	_ = database.Delete(&models.Otp{}, "email = ?", email)
	return true, nil
}