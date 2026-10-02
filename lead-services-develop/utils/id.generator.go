package utils

import (
	"math/rand"
	"time"

	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/models"
)

const charset = "0123456789"

func GenerateEmployeeID(db db.Database) string {
	rand.Seed(time.Now().UnixNano())

	for {
		empID := make([]byte, 6)
		for i := range empID {
			empID[i] = charset[rand.Intn(len(charset))]
		}
		empIDStr := string(empID)

		var count int64
		db.Model(&models.Agent{}).Where("id = ?", empIDStr).Count(&count) // make sure to check correct column!

		if count == 0 {
			return empIDStr
		}
	}
}
