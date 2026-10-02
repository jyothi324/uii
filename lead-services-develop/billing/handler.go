package billing

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/middleware"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func getService() (*Service, error) {
	gdb, err := db.GetControlPlaneDB()
	if err != nil {
		return nil, err
	}
	return NewService(gdb), nil
}

func resolveSubdomain(c *gin.Context, bodySubdomain string) string {
	if s := normalizeSubdomain(bodySubdomain); s != "" {
		return s
	}
	return normalizeSubdomain(middleware.GetTenant(c))
}

// hostSubdomain returns the subdomain exactly as the browser sees it ("appdev"),
// unlike middleware.GetTenant which strips env suffixes to key the tenant
// ("app"). Shown in the UI so dev is not mistaken for prod.
func hostSubdomain(c *gin.Context) string {
	// X-Tenant may be a bare identifier ("appdev") rather than a host.
	if xt := strings.TrimSpace(c.GetHeader("X-Tenant")); xt != "" &&
		!strings.Contains(xt, ".") && !strings.Contains(xt, "://") {
		if label := normalizeSubdomain(xt); label != "" && label != "www" && label != "default" {
			return label
		}
	}

	for _, raw := range []string{
		c.GetHeader("X-Tenant"),
		c.GetHeader("Referer"),
		c.GetHeader("Origin"),
		c.Request.Host,
	} {
		if sub := firstHostLabel(raw); sub != "" {
			return sub
		}
	}
	return ""
}

func firstHostLabel(raw string) string {
	host := strings.TrimSpace(raw)
	if host == "" {
		return ""
	}
	if idx := strings.Index(host, "://"); idx != -1 {
		host = host[idx+3:]
	}
	if idx := strings.IndexAny(host, "/?"); idx != -1 {
		host = host[:idx]
	}
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	parts := strings.Split(host, ".")
	// A bare "localhost" carries no tenant label.
	if len(parts) < 2 {
		return ""
	}
	label := normalizeSubdomain(parts[0])
	if label == "www" || label == "default" || isAllDigits(label) {
		return ""
	}
	return label
}

// isAllDigits filters out IP-address hosts (ALB health checks) whose first
// octet would otherwise look like a subdomain.
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// Signup POST /api/billing/signup
func (h *Handler) Signup(c *gin.Context) {
	svc, err := getService()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not available", "detail": err.Error()})
		return
	}
	var req SignupRequest
	_ = c.ShouldBindJSON(&req)
	if req.Subdomain == "" {
		req.Subdomain = middleware.GetTenant(c)
	}
	if normalizeSubdomain(req.Subdomain) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subdomain is required"})
		return
	}
	resp, err := svc.Signup(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CheckoutSession POST /api/billing/checkout-session
func (h *Handler) CheckoutSession(c *gin.Context) {
	svc, err := getService()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not available", "detail": err.Error()})
		return
	}
	var req CheckoutRequest
	_ = c.ShouldBindJSON(&req)
	sub := resolveSubdomain(c, req.Subdomain)
	if sub == "" || sub == "default" || sub == "www" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid subdomain is required"})
		return
	}
	resp, err := svc.CreateCheckoutSession(sub, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// OneTimeSession POST /api/billing/onetime-session
// Starts Checkout for a one-off advance / top-up charge (payment mode).
func (h *Handler) OneTimeSession(c *gin.Context) {
	svc, err := getService()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not available", "detail": err.Error()})
		return
	}
	var req OneTimeRequest
	_ = c.ShouldBindJSON(&req)
	sub := resolveSubdomain(c, req.Subdomain)
	if sub == "" || sub == "default" || sub == "www" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid subdomain is required"})
		return
	}
	resp, err := svc.CreateOneTimeSession(sub, req.Email, req.AmountCents)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PortalSession POST /api/billing/portal-session
func (h *Handler) PortalSession(c *gin.Context) {
	svc, err := getService()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not available", "detail": err.Error()})
		return
	}
	sub := resolveSubdomain(c, "")
	if sub == "" || sub == "default" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not resolve tenant subdomain"})
		return
	}
	resp, err := svc.CreatePortalSession(sub)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Status GET /api/billing/status
func (h *Handler) Status(c *gin.Context) {
	svc, err := getService()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not available", "detail": err.Error()})
		return
	}
	sub := resolveSubdomain(c, c.Query("subdomain"))
	if sub == "" || sub == "default" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not resolve tenant subdomain"})
		return
	}
	resp, err := svc.GetStatus(sub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp.HostSubdomain = hostSubdomain(c)
	c.JSON(http.StatusOK, resp)
}

type syncRequest struct {
	SessionID string `json:"session_id"`
}

// Sync POST /api/billing/sync
// Reconciles this tenant against Stripe and returns the refreshed status. Called
// when Checkout redirects back, so activation does not wait on a webhook.
func (h *Handler) Sync(c *gin.Context) {
	svc, err := getService()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not available", "detail": err.Error()})
		return
	}
	sub := resolveSubdomain(c, "")
	if sub == "" || sub == "default" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not resolve tenant subdomain"})
		return
	}
	var req syncRequest
	_ = c.ShouldBindJSON(&req)
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = strings.TrimSpace(c.Query("session_id"))
	}

	resp, err := svc.SyncFromStripe(sub, sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp.HostSubdomain = hostSubdomain(c)
	c.JSON(http.StatusOK, resp)
}

type stripeEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

// Webhook POST /api/billing/webhook
func (h *Handler) Webhook(c *gin.Context) {
	svc, err := getService()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not available"})
		return
	}

	const maxBody = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBody)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body failed"})
		return
	}

	secret := strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	sig := c.GetHeader("Stripe-Signature")

	// Fail closed: without a signing secret anyone could POST a forged event and
	// move a tenant's subscription status.
	if secret == "" {
		log.Printf("[billing] rejecting webhook: STRIPE_WEBHOOK_SECRET is not configured")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "webhook signing secret not configured"})
		return
	}
	if err := VerifyStripeSignature(payload, sig, secret); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature", "detail": err.Error()})
		return
	}

	var event stripeEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if svc.EventAlreadyProcessed(event.ID) {
		c.JSON(http.StatusOK, gin.H{"received": true, "duplicate": true})
		return
	}

	if err := handleStripeEvent(svc, event); err != nil {
		log.Printf("[billing] webhook handler error type=%s id=%s: %v", event.Type, event.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = svc.RecordEvent(event.ID, event.Type, string(payload))
	c.JSON(http.StatusOK, gin.H{"received": true})
}

func handleStripeEvent(svc *Service, event stripeEvent) error {
	switch event.Type {
	case "checkout.session.completed":
		var obj map[string]interface{}
		if err := json.Unmarshal(event.Data.Object, &obj); err != nil {
			return err
		}
		return handleCheckoutCompleted(svc, obj)
	case "customer.subscription.updated", "customer.subscription.created":
		var obj map[string]interface{}
		if err := json.Unmarshal(event.Data.Object, &obj); err != nil {
			return err
		}
		return handleSubscriptionChange(svc, obj)
	case "customer.subscription.deleted":
		var obj map[string]interface{}
		if err := json.Unmarshal(event.Data.Object, &obj); err != nil {
			return err
		}
		return handleSubscriptionDeleted(svc, obj)
	case "invoice.payment_failed":
		var obj map[string]interface{}
		if err := json.Unmarshal(event.Data.Object, &obj); err != nil {
			return err
		}
		return handleInvoiceFailed(svc, obj)
	case "invoice.paid":
		var obj map[string]interface{}
		if err := json.Unmarshal(event.Data.Object, &obj); err != nil {
			return err
		}
		return handleInvoicePaid(svc, obj)
	default:
		log.Printf("[billing] ignoring event type=%s", event.Type)
		return nil
	}
}

func metadataString(obj map[string]interface{}, key string) string {
	meta, _ := obj["metadata"].(map[string]interface{})
	if meta == nil {
		return ""
	}
	return asString(meta[key])
}

func customerIDFrom(obj map[string]interface{}) string {
	switch v := obj["customer"].(type) {
	case string:
		return v
	case map[string]interface{}:
		return asString(v["id"])
	default:
		return ""
	}
}

func subscriptionIDFrom(obj map[string]interface{}) string {
	switch v := obj["subscription"].(type) {
	case string:
		return v
	case map[string]interface{}:
		return asString(v["id"])
	default:
		return asString(obj["id"])
	}
}

func handleCheckoutCompleted(svc *Service, sess map[string]interface{}) error {
	tenant, ok := tenantForEvent(svc, sess)
	if !ok {
		return nil
	}

	// Payment-mode = one-time advance / top-up. Record it only; do not treat it
	// as activating a subscription.
	if isOneTimeSession(sess) {
		if err := svc.setStripeCustomer(tenant, customerIDFrom(sess)); err != nil {
			return err
		}
		return svc.ApplyOneTimeSession(tenant, sess)
	}

	if err := svc.UpdateTenantStatus(tenant.ID, StatusActive, customerIDFrom(sess)); err != nil {
		return err
	}

	subID := subscriptionIDFrom(sess)
	if subID == "" || strings.HasPrefix(subID, "cs_") {
		// checkout session id mistaken — subscription field empty
		if s, ok := sess["subscription"].(string); ok {
			subID = s
		} else {
			return nil
		}
	}
	if subID == "" {
		return nil
	}
	// The session carries only the subscription id, so read the subscription back
	// from Stripe for the price and period end the UI shows.
	if tenant.StripeCustomerID == "" {
		tenant.StripeCustomerID = customerIDFrom(sess)
	}
	if err := svc.reconcileTenant(tenant); err == nil {
		return nil
	}
	return svc.UpsertSubscriptionFromStripe(tenant.ID, subID, "", StatusActive, nil, false)
}

func handleSubscriptionChange(svc *Service, sub map[string]interface{}) error {
	tenant, ok := tenantForEvent(svc, sub)
	if !ok {
		return nil
	}
	_, err := svc.ApplySubscriptionObject(tenant.ID, customerIDFrom(sub), sub)
	return err
}

func handleSubscriptionDeleted(svc *Service, sub map[string]interface{}) error {
	tenant, ok := tenantForEvent(svc, sub)
	if !ok {
		return nil
	}
	if err := svc.UpdateTenantStatus(tenant.ID, StatusCanceled, ""); err != nil {
		return err
	}
	periodEnd := subscriptionPeriodEnd(sub)
	return svc.UpsertSubscriptionFromStripe(
		tenant.ID,
		asString(sub["id"]),
		"",
		StatusCanceled,
		periodEnd,
		true,
	)
}

func handleInvoiceFailed(svc *Service, inv map[string]interface{}) error {
	cid := customerIDFrom(inv)
	if cid == "" {
		return nil
	}
	tenant, err := svc.FindTenantByStripeCustomer(cid)
	if err != nil {
		return nil
	}
	return svc.UpdateTenantStatus(tenant.ID, StatusPastDue, "")
}

func handleInvoicePaid(svc *Service, inv map[string]interface{}) error {
	cid := customerIDFrom(inv)
	if cid == "" {
		return nil
	}
	tenant, err := svc.FindTenantByStripeCustomer(cid)
	if err != nil {
		return nil
	}
	if tenant.Status == StatusPastDue || tenant.Status == StatusPending {
		return svc.UpdateTenantStatus(tenant.ID, StatusActive, "")
	}
	return nil
}

// eventEnv reads the env stamp written by billingEnv() when the object was created.
func eventEnv(obj map[string]interface{}) string {
	return metadataString(obj, "env")
}

// foreignEvent reports an event created by another deployment sharing this Stripe
// account. Unstamped (legacy) objects are not treated as foreign.
func foreignEvent(obj map[string]interface{}) bool {
	ours, theirs := billingEnv(), eventEnv(obj)
	return ours != "" && theirs != "" && ours != theirs
}

// tenantForEvent maps an event onto a tenant, or reports that it should be
// skipped. Skipped events are acked rather than errored, so Stripe does not retry
// another environment's traffic forever.
func tenantForEvent(svc *Service, obj map[string]interface{}) (*Tenant, bool) {
	if foreignEvent(obj) {
		log.Printf("[billing] ignoring event from env=%s (this is %s)", eventEnv(obj), billingEnv())
		return nil, false
	}
	tenant, err := resolveTenant(
		svc,
		metadataString(obj, "tenant_id"),
		metadataString(obj, "subdomain"),
		customerIDFrom(obj),
		eventEnv(obj) == billingEnv(),
	)
	if err != nil {
		log.Printf("[billing] no tenant for event (subdomain=%q customer=%q): %v — skipping",
			metadataString(obj, "subdomain"), customerIDFrom(obj), err)
		return nil, false
	}
	return tenant, true
}

// resolveTenant maps a Stripe event onto a control-plane tenant. tenant_id and
// customer id are unique to this environment's control plane, so they are tried
// first. The subdomain fallback is only safe for events this environment created:
// with a shared Stripe account, dev's "appdev" and prod's "app" both normalize to
// "app", so matching on subdomain alone would let one environment rewrite the
// other's tenant.
func resolveTenant(svc *Service, tenantIDStr, subdomain, customerID string, subdomainOK bool) (*Tenant, error) {
	if tenantIDStr != "" {
		id, err := uuid.Parse(tenantIDStr)
		if err == nil {
			var t Tenant
			if err := svc.db.First(&t, "id = ?", id).Error; err == nil {
				return &t, nil
			}
		}
	}
	if customerID != "" {
		if t, err := svc.FindTenantByStripeCustomer(customerID); err == nil {
			return t, nil
		}
	}
	if subdomain != "" && subdomainOK {
		return svc.FindTenantBySubdomain(subdomain)
	}
	return nil, fmt.Errorf("cannot resolve tenant from Stripe event")
}
