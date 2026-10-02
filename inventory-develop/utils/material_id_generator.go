package utils

import (
	"inventory/db"
	"inventory/models"
	"math/rand"
)

const charset = "0123456789"

func GenerateEmployeeID(db db.Database) uint {
	for {
		id := uint(rand.Intn(900000) + 100000)

		var count int64
		db.Model(&models.Materials{}).
			Where("id = ?", id).
			Count(&count)

		if count == 0 {
			return id
		}
	}
}
