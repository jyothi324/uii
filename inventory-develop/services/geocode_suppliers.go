package services

import (
	"log"
	"time"

	"inventory/db"
	"inventory/models"
	"inventory/utils"
)

// geocodeExistingSuppliers backfills Lat/Lng for every supplier whose stored
// address doesn't have coordinates yet. Run once per tenant:
//
//	go run ./cmd geocode-suppliers default
//	go run ./cmd geocode-suppliers tenant_a tenant_b
//
// Tenant names resolve the same way GetTenantDB does elsewhere in this repo
// ("default" = the base DB_NAME database, anything else = "tenant_<name>").
func GeocodeExistingSuppliers(tenants []string) error {
	tm := db.GetTenantManager()

	for _, tenant := range tenants {
		database, err := tm.GetTenantDB(tenant)
		if err != nil {
			log.Printf("[%s] could not connect: %v", tenant, err)
			continue
		}

		var suppliers []models.Supplier
		if err := database.Find(&suppliers); err != nil {
			log.Printf("[%s] could not load suppliers: %v", tenant, err)
			continue
		}

		log.Printf("[%s] %d suppliers found, checking for missing coordinates", tenant, len(suppliers))

		for i := range suppliers {
			s := &suppliers[i]
			addr := s.Details.Address
			if addr.Lat != nil && addr.Lng != nil {
				continue // already geocoded
			}

			lat, lng, err := utils.GeocodeAddress(addr.Street, addr.City, addr.State, addr.Zip, addr.Country)
			if err != nil {
				log.Printf("[%s] supplier %d (%s): geocode failed: %v", tenant, s.ID, s.Name, err)
				continue
			}

			s.Details.Address.Lat = &lat
			s.Details.Address.Lng = &lng
			if err := database.Save(s); err != nil {
				log.Printf("[%s] supplier %d (%s): save failed: %v", tenant, s.ID, s.Name, err)
				continue
			}
			log.Printf("[%s] supplier %d (%s): geocoded to %.6f,%.6f", tenant, s.ID, s.Name, lat, lng)
			time.Sleep(200 * time.Millisecond)
		}
	}

	return nil
}
