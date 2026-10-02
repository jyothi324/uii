package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const stripeAPIBase = "https://api.stripe.com/v1"

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func stripeConfigured() bool {
	return strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")) != "" &&
		strings.TrimSpace(os.Getenv("STRIPE_PRICE_MONTHLY")) != ""
}

func (s *Service) ensureStripe() error {
	if strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")) == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is not configured")
	}
	if strings.TrimSpace(os.Getenv("STRIPE_PRICE_MONTHLY")) == "" {
		return fmt.Errorf("STRIPE_PRICE_MONTHLY is not configured")
	}
	return nil
}

// One-time (advance / top-up) payments are optional: without a one-time price the
// feature stays hidden and subscriptions keep working unchanged.
func oneTimeConfigured() bool {
	return strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")) != "" &&
		strings.TrimSpace(os.Getenv("STRIPE_PRICE_ONETIME")) != ""
}

func (s *Service) ensureOneTime() error {
	if strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")) == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is not configured")
	}
	if strings.TrimSpace(os.Getenv("STRIPE_PRICE_ONETIME")) == "" {
		return fmt.Errorf("one-time payments are not configured (STRIPE_PRICE_ONETIME)")
	}
	return nil
}

func stripeDo(method, path string, form url.Values) (map[string]interface{}, error) {
	key := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, stripeAPIBase+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(key, "")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("stripe decode: %w", err)
	}
	if resp.StatusCode >= 400 {
		msg := "stripe error"
		if e, ok := out["error"].(map[string]interface{}); ok {
			if m, ok := e["message"].(string); ok {
				msg = m
			}
		}
		return out, fmt.Errorf("%s", msg)
	}
	return out, nil
}

// billingEnv labels which deployment created a Stripe object. It is stamped into
// metadata so two environments can share one Stripe account without their
// webhooks writing over each other — dev's "appdev" and prod's "app" normalize to
// the same tenant key, so the env stamp is the only way to tell them apart.
func billingEnv() string {
	if v := strings.TrimSpace(os.Getenv("BILLING_ENV")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("APP_ENV"))
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// billingRedirectURL reads envKey (optional), applies {subdomain} placeholder,
// and falls back to defaultTpl when unset. Defaults use tenant.localhost on port 80.
func billingRedirectURL(envKey, subdomain, defaultTpl string) string {
	tpl := strings.TrimSpace(os.Getenv(envKey))
	if tpl == "" {
		tpl = defaultTpl
	}
	if tpl == "" {
		return ""
	}
	return strings.ReplaceAll(tpl, "{subdomain}", subdomain)
}

type SignupRequest struct {
	Subdomain   string `json:"subdomain"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type SignupResponse struct {
	TenantID  string `json:"tenant_id"`
	Subdomain string `json:"subdomain"`
	Status    string `json:"status"`
}

func normalizeSubdomain(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	var b strings.Builder
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func (s *Service) Signup(req SignupRequest) (*SignupResponse, error) {
	sub := normalizeSubdomain(req.Subdomain)
	if sub == "" || sub == "www" || sub == "default" {
		return nil, fmt.Errorf("invalid subdomain")
	}

	var existing Tenant
	err := s.db.Where("subdomain = ?", sub).First(&existing).Error
	if err == nil {
		return &SignupResponse{
			TenantID:  existing.ID.String(),
			Subdomain: existing.Subdomain,
			Status:    existing.Status,
		}, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	t := Tenant{
		Subdomain:   sub,
		DisplayName: strings.TrimSpace(req.DisplayName),
		Email:       strings.TrimSpace(req.Email),
		Status:      StatusPending,
	}
	if err := s.db.Create(&t).Error; err != nil {
		return nil, err
	}
	return &SignupResponse{
		TenantID:  t.ID.String(),
		Subdomain: t.Subdomain,
		Status:    t.Status,
	}, nil
}

type CheckoutRequest struct {
	Subdomain string `json:"subdomain"`
	Email     string `json:"email"`
}

// OneTimeRequest is a one-off advance / top-up. AmountCents is required so the
// payer chooses the charge — the catalog one-time price is only a feature flag /
// product reference, not a fixed amount.
type OneTimeRequest struct {
	Subdomain   string `json:"subdomain"`
	Email       string `json:"email"`
	AmountCents int64  `json:"amount_cents"`
}

type CheckoutResponse struct {
	URL       string `json:"url"`
	SessionID string `json:"session_id"`
}

func (s *Service) CreateCheckoutSession(subdomain, email string) (*CheckoutResponse, error) {
	if err := s.ensureStripe(); err != nil {
		return nil, err
	}

	sub := normalizeSubdomain(subdomain)
	if sub == "" {
		return nil, fmt.Errorf("subdomain is required")
	}

	var tenant Tenant
	if err := s.db.Where("subdomain = ?", sub).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found; call /api/billing/signup first")
		}
		return nil, err
	}

	if email == "" {
		email = tenant.Email
	}

	// Refuse a second Checkout for a tenant that is already subscribed, otherwise
	// clicking Subscribe again bills them on a duplicate subscription. Stripe is
	// the source of truth here because the local status may be a stale "pending".
	if tenant.StripeCustomerID != "" {
		if err := s.reconcileTenant(&tenant); err != nil {
			log.Printf("[billing] pre-checkout reconcile %s failed: %v", sub, err)
		}
		if IsSubscribedStatus(tenant.Status) {
			return nil, fmt.Errorf("this tenant already has a %s subscription — use Manage billing to change or cancel it", tenant.Status)
		}
	}

	if err := s.ensureCustomer(&tenant, email); err != nil {
		return nil, err
	}

	successURL := billingRedirectURL("STRIPE_SUCCESS_URL", sub,
		"http://{subdomain}.localhost/settings?billing=success&session_id={CHECKOUT_SESSION_ID}")
	cancelURL := billingRedirectURL("STRIPE_CANCEL_URL", sub,
		"http://{subdomain}.localhost/settings?billing=cancel")
	priceID := os.Getenv("STRIPE_PRICE_MONTHLY")

	form := url.Values{}
	form.Set("mode", "subscription")
	form.Set("customer", tenant.StripeCustomerID)
	form.Set("success_url", successURL)
	form.Set("cancel_url", cancelURL)
	form.Set("line_items[0][price]", priceID)
	form.Set("line_items[0][quantity]", "1")
	form.Set("payment_method_types[0]", "card")
	form.Set("metadata[subdomain]", sub)
	form.Set("metadata[tenant_id]", tenant.ID.String())
	form.Set("subscription_data[metadata][subdomain]", sub)
	form.Set("subscription_data[metadata][tenant_id]", tenant.ID.String())
	if env := billingEnv(); env != "" {
		form.Set("metadata[env]", env)
		form.Set("subscription_data[metadata][env]", env)
	}

	sess, err := stripeDo(http.MethodPost, "/checkout/sessions", form)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}

	return &CheckoutResponse{
		URL:       asString(sess["url"]),
		SessionID: asString(sess["id"]),
	}, nil
}

// ensureCustomer creates the tenant's Stripe customer on first purchase and
// stores the id, so subscriptions and one-time payments share one customer.
func (s *Service) ensureCustomer(tenant *Tenant, email string) error {
	if tenant.StripeCustomerID != "" {
		return nil
	}

	form := url.Values{}
	if email != "" {
		form.Set("email", email)
	}
	if tenant.DisplayName != "" {
		form.Set("name", tenant.DisplayName)
	}
	form.Set("metadata[subdomain]", tenant.Subdomain)
	form.Set("metadata[tenant_id]", tenant.ID.String())
	if env := billingEnv(); env != "" {
		form.Set("metadata[env]", env)
	}
	cust, err := stripeDo(http.MethodPost, "/customers", form)
	if err != nil {
		return fmt.Errorf("create stripe customer: %w", err)
	}

	tenant.StripeCustomerID = asString(cust["id"])
	if email != "" {
		tenant.Email = email
	}
	return s.db.Save(tenant).Error
}

// kindOneTime marks Stripe objects created by the pay-once (advance / top-up)
// flow, so a payment-mode session is never mistaken for starting a subscription.
const kindOneTime = "one_time"

// CreateOneTimeSession starts Checkout in payment mode for a one-off charge
// (advance, top-up, etc.). AmountCents is chosen by the payer. Allowed whether
// or not the tenant already has a monthly subscription.
func (s *Service) CreateOneTimeSession(subdomain, email string, amountCents int64) (*CheckoutResponse, error) {
	if err := s.ensureOneTime(); err != nil {
		return nil, err
	}
	// Stripe's card minimum for USD is $0.50; keep a sane upper bound too.
	if amountCents < 50 {
		return nil, fmt.Errorf("amount must be at least $0.50")
	}
	if amountCents > 1_000_000_00 { // $1,000,000.00
		return nil, fmt.Errorf("amount is too large")
	}

	sub := normalizeSubdomain(subdomain)
	if sub == "" {
		return nil, fmt.Errorf("subdomain is required")
	}

	var tenant Tenant
	if err := s.db.Where("subdomain = ?", sub).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found; call /api/billing/signup first")
		}
		return nil, err
	}

	if email == "" {
		email = tenant.Email
	}

	if err := s.ensureCustomer(&tenant, email); err != nil {
		return nil, err
	}

	currency, productID, err := oneTimePriceDefaults()
	if err != nil {
		return nil, err
	}

	successURL := billingRedirectURL("STRIPE_SUCCESS_URL", sub,
		"http://{subdomain}.localhost/settings?billing=success&session_id={CHECKOUT_SESSION_ID}")
	cancelURL := billingRedirectURL("STRIPE_CANCEL_URL", sub,
		"http://{subdomain}.localhost/settings?billing=cancel")

	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("customer", tenant.StripeCustomerID)
	form.Set("success_url", successURL)
	form.Set("cancel_url", cancelURL)
	// price_data (not a fixed price id) so the amount the user typed is what
	// Checkout charges — a catalog price would lock them to $5.
	form.Set("line_items[0][price_data][currency]", currency)
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(amountCents, 10))
	if productID != "" {
		form.Set("line_items[0][price_data][product]", productID)
	} else {
		form.Set("line_items[0][price_data][product_data][name]", "Tulip one-time payment")
	}
	form.Set("line_items[0][quantity]", "1")
	form.Set("payment_method_types[0]", "card")
	form.Set("metadata[subdomain]", sub)
	form.Set("metadata[tenant_id]", tenant.ID.String())
	form.Set("metadata[kind]", kindOneTime)
	form.Set("metadata[amount_cents]", strconv.FormatInt(amountCents, 10))
	form.Set("payment_intent_data[metadata][subdomain]", sub)
	form.Set("payment_intent_data[metadata][tenant_id]", tenant.ID.String())
	form.Set("payment_intent_data[metadata][kind]", kindOneTime)
	if env := billingEnv(); env != "" {
		form.Set("metadata[env]", env)
		form.Set("payment_intent_data[metadata][env]", env)
	}

	sess, err := stripeDo(http.MethodPost, "/checkout/sessions", form)
	if err != nil {
		return nil, fmt.Errorf("create one-time checkout session: %w", err)
	}

	return &CheckoutResponse{
		URL:       asString(sess["url"]),
		SessionID: asString(sess["id"]),
	}, nil
}

// oneTimePriceDefaults reads currency (and product, when present) from the
// configured one-time catalog price so ad-hoc Checkout amounts stay on the same
// product / currency as the Stripe catalog entry.
func oneTimePriceDefaults() (currency, productID string, err error) {
	priceID := strings.TrimSpace(os.Getenv("STRIPE_PRICE_ONETIME"))
	if priceID == "" {
		return "usd", "", nil
	}
	price, err := stripeDo(http.MethodGet, "/prices/"+url.PathEscape(priceID), nil)
	if err != nil {
		log.Printf("[billing] could not read STRIPE_PRICE_ONETIME %s: %v — defaulting to usd", priceID, err)
		return "usd", "", nil
	}
	currency = asString(price["currency"])
	if currency == "" {
		currency = "usd"
	}
	switch p := price["product"].(type) {
	case string:
		productID = p
	case map[string]interface{}:
		productID = asString(p["id"])
	}
	return currency, productID, nil
}

func isOneTimeSession(sess map[string]interface{}) bool {
	return asString(sess["mode"]) == "payment" || metadataString(sess, "kind") == kindOneTime
}

func paymentIntentIDFrom(obj map[string]interface{}) string {
	switch v := obj["payment_intent"].(type) {
	case string:
		return v
	case map[string]interface{}:
		return asString(v["id"])
	default:
		return ""
	}
}

// ApplyOneTimeSession records a paid one-time Checkout session. It does not
// change subscription status — this is an advance / top-up, not a plan change.
// Safe to call repeatedly: the session id is unique in one_time_payments.
func (s *Service) ApplyOneTimeSession(tenant *Tenant, sess map[string]interface{}) error {
	sessionID := asString(sess["id"])
	if sessionID == "" {
		return fmt.Errorf("checkout session has no id")
	}
	if st := asString(sess["payment_status"]); st != "paid" && st != "no_payment_required" {
		log.Printf("[billing] one-time session %s payment_status=%q — not recorded", sessionID, st)
		return nil
	}

	var existing OneTimePayment
	err := s.db.Where("stripe_session_id = ?", sessionID).First(&existing).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	amount, _ := sess["amount_total"].(float64)
	row := OneTimePayment{
		TenantID:              tenant.ID,
		StripeSessionID:       sessionID,
		StripePaymentIntentID: paymentIntentIDFrom(sess),
		StripePriceID:         "", // ad-hoc price_data; amount_total is the source of truth
		AmountTotal:           int64(amount),
		Currency:              asString(sess["currency"]),
	}
	if err := s.db.Create(&row).Error; err != nil {
		if s.db.Where("stripe_session_id = ?", sessionID).First(&existing).Error == nil {
			return nil
		}
		return err
	}

	log.Printf("[billing] one-time payment %s recorded for %s amount=%d %s",
		sessionID, tenant.Subdomain, row.AmountTotal, row.Currency)
	return nil
}

func (s *Service) setStripeCustomer(tenant *Tenant, customerID string) error {
	if customerID == "" || tenant.StripeCustomerID == customerID {
		return nil
	}
	if err := s.db.Model(&Tenant{}).Where("id = ?", tenant.ID).Updates(map[string]interface{}{
		"stripe_customer_id": customerID,
		"updated_at":         time.Now(),
	}).Error; err != nil {
		return err
	}
	tenant.StripeCustomerID = customerID
	return nil
}

func (s *Service) latestOneTimePayment(tenantID uuid.UUID) *OneTimePayment {
	var row OneTimePayment
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at desc").First(&row).Error; err != nil {
		return nil
	}
	return &row
}

func (s *Service) listOneTimePayments(tenantID uuid.UUID, limit int) []OneTimePayment {
	if limit <= 0 {
		limit = 25
	}
	var rows []OneTimePayment
	_ = s.db.Where("tenant_id = ?", tenantID).Order("created_at desc").Limit(limit).Find(&rows).Error
	return rows
}

const (
	PaymentKindOneTime      = "one_time"
	PaymentKindSubscription = "subscription"
)

// PaymentHistoryItem is one row in Settings → Billing payment history.
type PaymentHistoryItem struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"` // one_time | subscription
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	PaidAt      time.Time `json:"paid_at"`
	Reference   string    `json:"reference,omitempty"`
}

// buildPaymentHistory merges recorded one-time charges with Stripe subscription
// invoices, newest first. One-time rows win when both refer to the same payment.
func (s *Service) buildPaymentHistory(tenant *Tenant) []PaymentHistoryItem {
	const limit = 25
	seenPI := map[string]bool{}
	out := make([]PaymentHistoryItem, 0, limit)

	for _, p := range s.listOneTimePayments(tenant.ID, limit) {
		if p.StripePaymentIntentID != "" {
			seenPI[p.StripePaymentIntentID] = true
		}
		ref := p.StripePaymentIntentID
		if ref == "" {
			ref = p.StripeSessionID
		}
		out = append(out, PaymentHistoryItem{
			ID:          p.ID.String(),
			Kind:        PaymentKindOneTime,
			Description: "One-time payment",
			Amount:      p.AmountTotal,
			Currency:    p.Currency,
			Status:      "paid",
			PaidAt:      p.CreatedAt,
			Reference:   ref,
		})
	}

	if tenant.StripeCustomerID != "" && stripeConfigured() {
		invoices, err := fetchCustomerInvoices(tenant.StripeCustomerID, limit)
		if err != nil {
			log.Printf("[billing] list invoices for %s: %v", tenant.Subdomain, err)
		} else {
			for _, inv := range invoices {
				pi := paymentIntentIDFrom(inv)
				if pi != "" && seenPI[pi] {
					continue
				}
				status := asString(inv["status"])
				if status == "" || status == "draft" || status == "void" {
					continue
				}
				amount, _ := inv["amount_paid"].(float64)
				if amount <= 0 {
					if total, ok := inv["amount_due"].(float64); ok {
						amount = total
					}
				}
				paidAt := time.Now()
				if v, ok := inv["status_transitions"].(map[string]interface{}); ok {
					if paid, ok := v["paid_at"].(float64); ok && paid > 0 {
						paidAt = time.Unix(int64(paid), 0)
					}
				} else if v, ok := inv["created"].(float64); ok && v > 0 {
					paidAt = time.Unix(int64(v), 0)
				}
				desc := asString(inv["description"])
				if desc == "" {
					desc = "Subscription"
					if period := invoicePeriodLabel(inv); period != "" {
						desc = "Subscription · " + period
					}
				}
				out = append(out, PaymentHistoryItem{
					ID:          asString(inv["id"]),
					Kind:        PaymentKindSubscription,
					Description: desc,
					Amount:      int64(amount),
					Currency:    asString(inv["currency"]),
					Status:      status,
					PaidAt:      paidAt,
					Reference:   asString(inv["id"]),
				})
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].PaidAt.After(out[j].PaidAt)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func fetchCustomerInvoices(customerID string, limit int) ([]map[string]interface{}, error) {
	q := url.Values{}
	q.Set("customer", customerID)
	q.Set("limit", strconv.Itoa(limit))
	q.Add("expand[]", "data.lines")
	out, err := stripeDo(http.MethodGet, "/invoices?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	data, _ := out["data"].([]interface{})
	invoices := make([]map[string]interface{}, 0, len(data))
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			invoices = append(invoices, m)
		}
	}
	return invoices, nil
}

func invoicePeriodLabel(inv map[string]interface{}) string {
	lines, _ := inv["lines"].(map[string]interface{})
	data, _ := lines["data"].([]interface{})
	if len(data) == 0 {
		return ""
	}
	line, _ := data[0].(map[string]interface{})
	period, _ := line["period"].(map[string]interface{})
	if period == nil {
		return ""
	}
	start, _ := period["start"].(float64)
	end, _ := period["end"].(float64)
	if start <= 0 || end <= 0 {
		return ""
	}
	from := time.Unix(int64(start), 0).Format("Jan 2")
	to := time.Unix(int64(end), 0).Format("Jan 2, 2006")
	return from + " – " + to
}

type PortalResponse struct {
	URL string `json:"url"`
}

func (s *Service) CreatePortalSession(subdomain string) (*PortalResponse, error) {
	if err := s.ensureStripe(); err != nil {
		return nil, err
	}

	sub := normalizeSubdomain(subdomain)
	var tenant Tenant
	if err := s.db.Where("subdomain = ?", sub).First(&tenant).Error; err != nil {
		return nil, fmt.Errorf("tenant not found")
	}
	if tenant.StripeCustomerID == "" {
		return nil, fmt.Errorf("no Stripe customer for this tenant; subscribe first")
	}

	returnURL := billingRedirectURL("STRIPE_PORTAL_RETURN_URL", sub, "")
	if returnURL == "" {
		returnURL = billingRedirectURL("STRIPE_CANCEL_URL", sub,
			"http://{subdomain}.localhost/settings")
	}

	form := url.Values{}
	form.Set("customer", tenant.StripeCustomerID)
	form.Set("return_url", returnURL)
	sess, err := stripeDo(http.MethodPost, "/billing_portal/sessions", form)
	if err != nil {
		return nil, fmt.Errorf("create portal session: %w", err)
	}
	return &PortalResponse{URL: asString(sess["url"])}, nil
}

type StatusResponse struct {
	Configured bool   `json:"configured"`
	Subdomain  string `json:"subdomain"`
	// HostSubdomain is the subdomain as it appears in the browser ("appdev"),
	// before the env suffix is stripped to form Subdomain ("app").
	HostSubdomain        string     `json:"host_subdomain,omitempty"`
	DisplayName          string     `json:"display_name,omitempty"`
	Status               string     `json:"status"`
	StripeCustomerID     string     `json:"stripe_customer_id,omitempty"`
	SubscriptionStatus   string     `json:"subscription_status,omitempty"`
	StripeSubscriptionID string     `json:"stripe_subscription_id,omitempty"`
	StripePriceID        string     `json:"stripe_price_id,omitempty"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end"`
	HasSubscription      bool       `json:"has_subscription"`
	// OneTimeAvailable is true when STRIPE_PRICE_ONETIME is set. One-time charges
	// (advance / top-up) are independent of the monthly subscription.
	OneTimeAvailable bool `json:"one_time_available"`
	// Last one-time payment, if any — so the UI can confirm a recent charge.
	LastOneTimeAt       *time.Time `json:"last_one_time_at,omitempty"`
	LastOneTimeAmount   int64      `json:"last_one_time_amount,omitempty"`
	LastOneTimeCurrency string     `json:"last_one_time_currency,omitempty"`
	// JustRecordedOneTime is set on a sync response when this request applied a
	// one-time Checkout session, so the UI can toast correctly.
	JustRecordedOneTime bool `json:"just_recorded_one_time,omitempty"`
	// Payments is recent one-time charges plus Stripe subscription invoices.
	Payments []PaymentHistoryItem `json:"payments,omitempty"`
}

// IsSubscribedStatus reports whether a status means the tenant currently has a
// subscription with Stripe (past_due included — it is a payment problem on an
// existing subscription, not a reason to start a second one).
func IsSubscribedStatus(status string) bool {
	switch status {
	case StatusActive, StatusTrialing, StatusPastDue:
		return true
	default:
		return false
	}
}

func (s *Service) GetStatus(subdomain string) (*StatusResponse, error) {
	return s.status(subdomain, true)
}

// status builds the response, self-healing from Stripe first when the stored state
// cannot be trusted.
func (s *Service) status(subdomain string, allowSync bool) (*StatusResponse, error) {
	sub := normalizeSubdomain(subdomain)
	resp := &StatusResponse{
		Configured:       stripeConfigured(),
		Subdomain:        sub,
		Status:           "unknown",
		OneTimeAvailable: oneTimeConfigured(),
	}

	var tenant Tenant
	if err := s.db.Where("subdomain = ?", sub).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			resp.Status = "not_registered"
			return resp, nil
		}
		return nil, err
	}

	subRow := s.latestSubscription(tenant.ID)

	if allowSync && stripeConfigured() && tenant.StripeCustomerID != "" && needsReconcile(&tenant, subRow) {
		if err := s.reconcileTenant(&tenant); err != nil {
			log.Printf("[billing] reconcile %s failed: %v", sub, err)
		} else {
			subRow = s.latestSubscription(tenant.ID)
		}
	}

	resp.DisplayName = tenant.DisplayName
	resp.Status = tenant.Status
	resp.StripeCustomerID = tenant.StripeCustomerID

	if subRow != nil {
		resp.HasSubscription = true
		resp.SubscriptionStatus = subRow.Status
		resp.StripeSubscriptionID = subRow.StripeSubscriptionID
		resp.StripePriceID = subRow.StripePriceID
		resp.CurrentPeriodEnd = subRow.CurrentPeriodEnd
		resp.CancelAtPeriodEnd = subRow.CancelAtPeriodEnd
	}

	if ot := s.latestOneTimePayment(tenant.ID); ot != nil {
		t := ot.CreatedAt
		resp.LastOneTimeAt = &t
		resp.LastOneTimeAmount = ot.AmountTotal
		resp.LastOneTimeCurrency = ot.Currency
	}

	resp.Payments = s.buildPaymentHistory(&tenant)

	return resp, nil
}

func (s *Service) latestSubscription(tenantID uuid.UUID) *Subscription {
	var row Subscription
	if err := s.db.Where("tenant_id = ?", tenantID).Order("updated_at desc").First(&row).Error; err != nil {
		return nil
	}
	return &row
}

// needsReconcile reports when the control plane cannot be trusted without asking
// Stripe: a tenant that looks unsubscribed but already has a customer (a webhook
// that never arrived), one that looks subscribed with no subscription row, or a
// stored period that has already elapsed — a renewal or cancellation nobody told
// us about.
func needsReconcile(tenant *Tenant, subRow *Subscription) bool {
	if !IsSubscribedStatus(tenant.Status) {
		return true
	}
	if subRow == nil {
		return true
	}
	return subRow.CurrentPeriodEnd != nil && subRow.CurrentPeriodEnd.Before(time.Now())
}

// SyncFromStripe reconciles the control plane with Stripe and returns the fresh
// status. Webhooks remain the primary path; this covers a missed or delayed one
// and is what the UI calls when Checkout redirects back, so the tenant does not
// sit on "pending" while waiting for an event.
func (s *Service) SyncFromStripe(subdomain, sessionID string) (*StatusResponse, error) {
	if err := s.ensureStripe(); err != nil {
		return nil, err
	}

	sub := normalizeSubdomain(subdomain)
	var tenant Tenant
	if err := s.db.Where("subdomain = ?", sub).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.status(sub, false)
		}
		return nil, err
	}

	// A Checkout redirect carries the session id, which is the only link to the
	// customer when the tenant row was created before any Stripe customer.
	recordedOneTime := false
	if sessionID != "" {
		sess, err := stripeDo(http.MethodGet, "/checkout/sessions/"+url.PathEscape(sessionID), nil)
		if err != nil {
			log.Printf("[billing] fetch checkout session %s: %v", sessionID, err)
		} else {
			if cust := customerIDFrom(sess); cust != "" && cust != tenant.StripeCustomerID {
				tenant.StripeCustomerID = cust
				if err := s.db.Save(&tenant).Error; err != nil {
					return nil, err
				}
			}
			// Payment-mode sessions leave no subscription for reconcile to find.
			if isOneTimeSession(sess) {
				if err := s.ApplyOneTimeSession(&tenant, sess); err != nil {
					return nil, err
				}
				recordedOneTime = true
			}
		}
	}

	if err := s.reconcileTenant(&tenant); err != nil {
		return nil, err
	}
	resp, err := s.status(sub, false)
	if err != nil {
		return nil, err
	}
	resp.JustRecordedOneTime = recordedOneTime
	return resp, nil
}

// reconcileTenant pulls the tenant's subscriptions from Stripe and writes the
// most relevant one onto the control plane.
func (s *Service) reconcileTenant(tenant *Tenant) error {
	if tenant.StripeCustomerID == "" {
		return nil
	}
	subs, err := fetchCustomerSubscriptions(tenant.StripeCustomerID)
	if err != nil {
		return err
	}
	best := pickBestSubscription(subs)
	if best == nil {
		return nil
	}
	status, err := s.ApplySubscriptionObject(tenant.ID, tenant.StripeCustomerID, best)
	if err != nil {
		return err
	}
	tenant.Status = status
	return nil
}

// ApplySubscriptionObject writes a Stripe subscription object onto the tenant and
// its subscription row, returning the mapped status.
func (s *Service) ApplySubscriptionObject(tenantID uuid.UUID, customerID string, sub map[string]interface{}) (string, error) {
	status := MapStripeSubStatus(asString(sub["status"]))
	if err := s.UpdateTenantStatus(tenantID, status, customerID); err != nil {
		return "", err
	}
	return status, s.UpsertSubscriptionFromStripe(
		tenantID,
		asString(sub["id"]),
		subscriptionPriceID(sub),
		status,
		subscriptionPeriodEnd(sub),
		subscriptionCancelAtPeriodEnd(sub),
	)
}

func fetchCustomerSubscriptions(customerID string) ([]map[string]interface{}, error) {
	q := url.Values{}
	q.Set("customer", customerID)
	q.Set("status", "all")
	q.Set("limit", "10")
	out, err := stripeDo(http.MethodGet, "/subscriptions?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("list stripe subscriptions: %w", err)
	}
	data, _ := out["data"].([]interface{})
	subs := make([]map[string]interface{}, 0, len(data))
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			subs = append(subs, m)
		}
	}
	return subs, nil
}

// subscriptionRank orders Stripe statuses so a live subscription always wins over
// a leftover canceled one when a customer has several.
func subscriptionRank(status string) int {
	switch status {
	case "active", "trialing":
		return 4
	case "past_due", "unpaid":
		return 3
	case "incomplete", "paused":
		return 2
	case "incomplete_expired":
		return 1
	default: // canceled
		return 0
	}
}

func pickBestSubscription(subs []map[string]interface{}) map[string]interface{} {
	var best map[string]interface{}
	bestRank, bestCreated := -1, float64(-1)
	for _, sub := range subs {
		rank := subscriptionRank(asString(sub["status"]))
		created, _ := sub["created"].(float64)
		if rank > bestRank || (rank == bestRank && created > bestCreated) {
			best, bestRank, bestCreated = sub, rank, created
		}
	}
	return best
}

func subscriptionItems(sub map[string]interface{}) []interface{} {
	items, ok := sub["items"].(map[string]interface{})
	if !ok {
		return nil
	}
	data, _ := items["data"].([]interface{})
	return data
}

// subscriptionPeriodEnd reads current_period_end, which recent Stripe API
// versions moved from the subscription onto its items.
func subscriptionPeriodEnd(sub map[string]interface{}) *time.Time {
	if v, ok := sub["current_period_end"].(float64); ok && v > 0 {
		t := time.Unix(int64(v), 0)
		return &t
	}
	for _, item := range subscriptionItems(sub) {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := m["current_period_end"].(float64); ok && v > 0 {
			t := time.Unix(int64(v), 0)
			return &t
		}
	}
	return nil
}

func subscriptionPriceID(sub map[string]interface{}) string {
	for _, item := range subscriptionItems(sub) {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		switch price := m["price"].(type) {
		case map[string]interface{}:
			return asString(price["id"])
		case string:
			return price
		}
	}
	return ""
}

func subscriptionCancelAtPeriodEnd(sub map[string]interface{}) bool {
	v, _ := sub["cancel_at_period_end"].(bool)
	return v
}

func (s *Service) UpsertSubscriptionFromStripe(
	tenantID uuid.UUID,
	stripeSubID, priceID, status string,
	periodEnd *time.Time,
	cancelAtPeriodEnd bool,
) error {
	var existing Subscription
	err := s.db.Where("stripe_subscription_id = ?", stripeSubID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		row := Subscription{
			TenantID:             tenantID,
			StripeSubscriptionID: stripeSubID,
			StripePriceID:        priceID,
			Status:               status,
			CurrentPeriodEnd:     periodEnd,
			CancelAtPeriodEnd:    cancelAtPeriodEnd,
		}
		return s.db.Create(&row).Error
	}
	if err != nil {
		return err
	}
	existing.StripePriceID = priceID
	existing.Status = status
	existing.CurrentPeriodEnd = periodEnd
	existing.CancelAtPeriodEnd = cancelAtPeriodEnd
	return s.db.Save(&existing).Error
}

func (s *Service) UpdateTenantStatus(tenantID uuid.UUID, status, stripeCustomerID string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if stripeCustomerID != "" {
		updates["stripe_customer_id"] = stripeCustomerID
	}
	return s.db.Model(&Tenant{}).Where("id = ?", tenantID).Updates(updates).Error
}

func (s *Service) FindTenantBySubdomain(sub string) (*Tenant, error) {
	var t Tenant
	if err := s.db.Where("subdomain = ?", normalizeSubdomain(sub)).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Service) FindTenantByStripeCustomer(customerID string) (*Tenant, error) {
	var t Tenant
	if err := s.db.Where("stripe_customer_id = ?", customerID).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Service) EventAlreadyProcessed(eventID string) bool {
	var count int64
	s.db.Model(&BillingEvent{}).Where("stripe_event_id = ?", eventID).Count(&count)
	return count > 0
}

func (s *Service) RecordEvent(eventID, eventType, payload string) error {
	return s.db.Create(&BillingEvent{
		StripeEventID: eventID,
		Type:          eventType,
		Payload:       payload,
		ProcessedAt:   time.Now(),
	}).Error
}

func MapStripeSubStatus(status string) string {
	switch status {
	case "active":
		return StatusActive
	case "trialing":
		return StatusTrialing
	case "past_due":
		return StatusPastDue
	case "canceled", "unpaid":
		return StatusCanceled
	default:
		if status == "" {
			return StatusActive
		}
		return status
	}
}

// VerifyStripeSignature validates the Stripe-Signature header (t=…,v1=…).
func VerifyStripeSignature(payload []byte, header, secret string) error {
	if secret == "" {
		return fmt.Errorf("webhook secret empty")
	}
	var timestamp string
	var signatures []string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signatures = append(signatures, kv[1])
		}
	}
	if timestamp == "" || len(signatures) == 0 {
		return fmt.Errorf("invalid Stripe-Signature header")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	for _, sig := range signatures {
		if hmac.Equal([]byte(expected), []byte(sig)) {
			return nil
		}
	}
	return fmt.Errorf("signature mismatch")
}
