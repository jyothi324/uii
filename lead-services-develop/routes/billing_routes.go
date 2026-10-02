package routes

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/billing"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/middleware"
)

// RegisterBillingRoutes mounts NEW /api/billing/* endpoints only.
// Does not modify existing route handlers. Call before api.Use(AuthMiddleware)
// so public billing routes stay public; auth is applied to the subgroup only.
func RegisterBillingRoutes(api *gin.RouterGroup) {
	gdb, err := db.GetControlPlaneDB()
	if err != nil {
		log.Printf("[billing] skipping route registration: %v", err)
		// Still register handlers so clients get structured 503 from handler layer
	} else if err := billing.MigrateControlPlane(gdb); err != nil {
		log.Printf("[billing] control-plane migrate failed: %v", err)
	}

	h := billing.NewHandler()

	billingGroup := api.Group("/billing")
	{
		billingGroup.POST("/webhook", h.Webhook)
		billingGroup.POST("/signup", h.Signup)
		billingGroup.POST("/checkout-session", h.CheckoutSession)
		billingGroup.POST("/onetime-session", h.OneTimeSession)

		authBilling := billingGroup.Group("/")
		authBilling.Use(middleware.AuthMiddleware)
		{
			authBilling.GET("/status", h.Status)
			authBilling.POST("/portal-session", h.PortalSession)
			authBilling.POST("/sync", h.Sync)
		}
	}
	log.Printf("[billing] routes registered under /api/billing")
}
