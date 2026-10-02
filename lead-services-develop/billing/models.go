package billing

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tenant is a SaaS customer (subdomain) in the control-plane DB.
type Tenant struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Subdomain        string    `gorm:"uniqueIndex;size:64;not null" json:"subdomain"`
	DisplayName      string    `gorm:"size:255" json:"display_name"`
	Email            string    `gorm:"size:255" json:"email"`
	Status           string    `gorm:"size:32;not null;index" json:"status"` // pending|trialing|active|past_due|canceled|suspended
	StripeCustomerID string    `gorm:"size:128" json:"stripe_customer_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (Tenant) TableName() string { return "tenants" }

func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// Subscription tracks the Stripe subscription for a tenant.
type Subscription struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID             uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	StripeSubscriptionID string     `gorm:"uniqueIndex;size:128;not null" json:"stripe_subscription_id"`
	StripePriceID        string     `gorm:"size:128" json:"stripe_price_id"`
	Status               string     `gorm:"size:32;not null;index" json:"status"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end"`
	CancelAtPeriodEnd    bool       `gorm:"default:false" json:"cancel_at_period_end"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (Subscription) TableName() string { return "subscriptions" }

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// OneTimePayment records an advance / top-up charge made through Checkout in
// payment mode. It does not replace the monthly subscription — it sits beside
// it. The unique Checkout session id makes recording idempotent across webhook
// retries and the post-checkout sync.
type OneTimePayment struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID              uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	StripeSessionID       string    `gorm:"uniqueIndex;size:128;not null" json:"stripe_session_id"`
	StripePaymentIntentID string    `gorm:"size:128" json:"stripe_payment_intent_id"`
	StripePriceID         string    `gorm:"size:128" json:"stripe_price_id"`
	AmountTotal           int64     `json:"amount_total"`
	Currency              string    `gorm:"size:8" json:"currency"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (OneTimePayment) TableName() string { return "one_time_payments" }

func (p *OneTimePayment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// BillingEvent stores processed Stripe webhook events for idempotency.
type BillingEvent struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StripeEventID string    `gorm:"uniqueIndex;size:128;not null" json:"stripe_event_id"`
	Type          string    `gorm:"size:128;not null" json:"type"`
	Payload       string    `gorm:"type:text" json:"payload"`
	ProcessedAt   time.Time `json:"processed_at"`
}

func (BillingEvent) TableName() string { return "billing_events" }

const (
	StatusPending   = "pending"
	StatusTrialing  = "trialing"
	StatusActive    = "active"
	StatusPastDue   = "past_due"
	StatusCanceled  = "canceled"
	StatusSuspended = "suspended"
)

// MigrateControlPlane runs AutoMigrate for billing tables.
func MigrateControlPlane(db *gorm.DB) error {
	return db.AutoMigrate(&Tenant{}, &Subscription{}, &OneTimePayment{}, &BillingEvent{})
}
