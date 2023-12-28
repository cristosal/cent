package pay

import (
	"time"
)

type PricingSchedule = string

const (
	PricingAnnual  PricingSchedule = "annual"
	PricingMonthly PricingSchedule = "monthly"
	PricingOnce    PricingSchedule = "once"
)

type Price struct {
	ID         int64
	PlanID     int64
	Provider   string
	ProviderID string
	Amount     int64
	Currency   string
	Schedule   PricingSchedule
	TrialDays  int
}

func (p *Price) TableName() string {
	return "pay.price"
}

func (p *Price) HasTrial() bool {
	return p.TrialDays > 0
}

// TrialEnd returns the time at which the trial would end if it started now
func (p *Price) TrialEnd() time.Time {
	return time.Now().Add(time.Hour * 24 * time.Duration(p.TrialDays))
}

// Plan that customers will subscribe to
type Plan struct {
	ID          int64
	Name        string
	Description string
	Provider    string
	ProviderID  string
	Active      bool
}

func (p *Plan) TableName() string {
	return "pay.plan"
}

// Customer from a provider like stripe or paypal
type Customer struct {
	ID         int64  // internal (to this service)
	ProviderID string // external providers id
	Provider   string // the provider for this customer
	Name       string // customers name
	Email      string // customers email
}

func (c *Customer) TableName() string {
	return "pay.customer"
}

// Subscription represents a customers subscription to a Plan
type Subscription struct {
	ID         int64
	Provider   string
	ProviderID string
	CustomerID int64
	PriceID    int64
	Active     bool
	CreatedAt  time.Time
}

func (s *Subscription) TableName() string {
	return "pay.subscription"
}

type WebhookEvent struct {
	ID         int64
	Provider   string
	ProviderID string
	EventType  string
	Payload    []byte
}

func (e *WebhookEvent) TableName() string {
	return "pay.webhook_event"
}

type SubscriptionUser struct {
	SubscriptionID int64
	Username       string
}

func (SubscriptionUser) TableName() string {
	return "pay.subscription_user"
}
