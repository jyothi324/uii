package services

import (
	"log"
	"time"

	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/utils"
)

// GeocodeExistingCustomers backfills Lat/Lng inside Customer.Address
// (falling back to BillingAddress) for every customer whose stored address
// doesn't have coordinates yet. Run once per tenant:
//
//	go run ./cmd geocode-customers default
func GeocodeExistingCustomers(tenants []string) error {
	tm := db.GetTenantManager()

	for _, tenant := range tenants {
		database, err := tm.GetTenantDB(tenant)
		if err != nil {
			log.Printf("[%s] could not connect: %v", tenant, err)
			continue
		}

		var customers []models.Customer
		if err := database.Find(&customers); err != nil {
			log.Printf("[%s] could not load customers: %v", tenant, err)
			continue
		}

		log.Printf("[%s] %d customers found, checking for missing coordinates", tenant, len(customers))

		firstNonEmpty := func(a, b string) string {
			if a != "" {
				return a
			}
			return b
		}

		for i := range customers {
			c := &customers[i]

			// Skip if we already have coordinates on the primary Address —
			// that's the one the frontend reads and the one we write to.
			if c.Address.Lat != nil && c.Address.Lng != nil {
				continue
			}

			// Merge field-by-field: prefer whatever's in Address, fall back
			// to the matching field in BillingAddress. This handles partial
			// records (e.g. Street set but City missing, or vice versa)
			// instead of requiring one whole address object to be complete.
			street := firstNonEmpty(c.Address.Street, c.BillingAddress.Street)
			city := firstNonEmpty(c.Address.City, c.BillingAddress.City)
			state := firstNonEmpty(c.Address.State, c.BillingAddress.State)
			zip := firstNonEmpty(c.Address.Zip, c.BillingAddress.Zip)
			country := firstNonEmpty(c.Address.Country, c.BillingAddress.Country)

			if street == "" && city == "" && state == "" && zip == "" {
				log.Printf("[%s] customer %s (%s): no address data anywhere, skipping", tenant, c.ID, c.Name)
				continue
			}

			lat, lng, err := utils.GeocodeAddress(street, city, state, zip, country)
			if err != nil {
				log.Printf("[%s] customer %s (%s): geocode failed for \"%s, %s, %s %s, %s\": %v",
					tenant, c.ID, c.Name, street, city, state, zip, country, err)
				continue
			}

			// Always write onto the primary Address, regardless of which
			// fields came from BillingAddress — that keeps the frontend's
			// `c.address ?? c.billingAddress` read simple and consistent.
			c.Address.Lat = &lat
			c.Address.Lng = &lng
			if err := database.Save(c); err != nil {
				log.Printf("[%s] customer %s (%s): save failed: %v", tenant, c.ID, c.Name, err)
				continue
			}
			log.Printf("[%s] customer %s (%s): geocoded to %.6f,%.6f", tenant, c.ID, c.Name, lat, lng)
			time.Sleep(200 * time.Millisecond)
		}
	}

	return nil
}
