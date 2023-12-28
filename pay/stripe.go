package pay

import (
	"errors"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

const ProviderStripe = "stripe"

var ErrCheckoutFailed = errors.New("checkout failed")

type (
	// StripeConfig configures StripeService with necessary credentials and callbacks
	StripeConfig struct {
		Repo          *Repo
		Key           string
		WebhookSecret string
	}

	// StripeProvider interfaces with stripe for customer, plan and subscription data
	StripeProvider struct {
		*Repo
		config *StripeConfig
	}
)

// NewStripeProvider creates a provider service for interacting with stripe
func NewStripeProvider(config *StripeConfig) *StripeProvider {
	if config == nil {
		config = new(StripeConfig)
	}

	stripe.Key = config.Key
	return &StripeProvider{
		Repo:   config.Repo,
		config: config,
	}
}

// AddPlan directly in stripe
func (s *StripeProvider) AddPlan(p *Plan) error {
	_, err := product.New(&stripe.ProductParams{
		Name:        stripe.String(p.Name),
		Description: stripe.String(p.Description),
		Active:      stripe.Bool(p.Active),
	})
	return err
}

// UpdatePlan in stripe
func (s *StripeProvider) UpdatePlan(p *Plan) error {
	_, err := product.Update(p.ProviderID, &stripe.ProductParams{
		Name:        stripe.String(p.Name),
		Description: stripe.String(p.Description),
		Active:      stripe.Bool(p.Active),
	})

	return err
}

// RemovePlan from stripe
func (s *StripeProvider) RemovePlanByProviderID(providerID string) error {
	_, err := product.Del(providerID, nil)
	return err
}

// AddPrice directly in stripe
func (s *StripeProvider) AddPrice(p *Price) error {
	var sched string
	switch p.Schedule {
	case PricingAnnual:
		sched = string(stripe.PriceRecurringIntervalYear)
	case PricingMonthly:
		sched = string(stripe.PriceRecurringIntervalMonth)
	}

	pl, err := s.GetPlanByID(p.PlanID)
	if err != nil {
		return fmt.Errorf("plan with id %d not found", p.PlanID)
	}

	_, err = price.New(&stripe.PriceParams{
		Currency:   stripe.String(p.Currency),
		UnitAmount: stripe.Int64(p.Amount),
		Product:    stripe.String(pl.ProviderID),
		Recurring: &stripe.PriceRecurringParams{
			Interval:        stripe.String(sched),
			TrialPeriodDays: stripe.Int64(int64(p.TrialDays)),
			IntervalCount:   stripe.Int64(1),
		},
	})

	return err
}

// AddCustomer directly in stripe
func (s *StripeProvider) AddCustomer(c *Customer) error {
	_, err := customer.New(&stripe.CustomerParams{
		Name:  stripe.String(c.Name),
		Email: stripe.String(c.Email),
	})
	return err
}

// Update Customer directly in stripe
func (s *StripeProvider) UpdateCustomer(c *Customer) error {
	if c.ProviderID == "" {
		return errors.New("missing customer provider id")
	}

	_, err := customer.Update(c.ProviderID, &stripe.CustomerParams{
		Name:  stripe.String(c.Name),
		Email: stripe.String(c.Email),
	})

	return err
}

// RemoveCustomer directly in stripe
func (s *StripeProvider) RemoveCustomerByProviderID(providerID string) error {
	_, err := customer.Del(providerID, nil)
	return err
}

// Verify that the checkout was completed
func (s *StripeProvider) VerifyCheckout(sessionID string) error {
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return err
	}

	if sess.PaymentStatus == stripe.CheckoutSessionPaymentStatusUnpaid {
		return ErrCheckoutFailed
	}

	return nil
}

// CheckoutRequest
type CheckoutRequest struct {
	CustomerID  int64
	PriceID     int64
	RedirectURL string
}

// Checkout returns the url that a user has to visit in order to complete payment
// it registers the customer if it was unavailable
func (s *StripeProvider) Checkout(request *CheckoutRequest) (url string, err error) {
	customer, err := s.GetCustomerByID(request.CustomerID)
	if err != nil {
		return
	}

	price, err := s.GetPriceByID(request.PriceID)
	if err != nil {
		return
	}

	var trialEnd *int64 = nil

	if price.TrialDays > 0 {
		// we add one day of grace so that stripe displays the correct amount.
		// since trial end is calculated from current time, being one second off will result in days -1 being displayed in stripe checkout
		trialEnd = stripe.Int64(price.TrialEnd().Add(time.Hour * 24).Unix())
	}

	params := &stripe.CheckoutSessionParams{
		Customer:                stripe.String(customer.ProviderID),
		SuccessURL:              stripe.String(request.RedirectURL),
		Mode:                    stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentMethodCollection: stripe.String("if_required"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(price.ProviderID),
				Quantity: stripe.Int64(1),
			},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialEnd: trialEnd,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return
	}

	url = sess.URL
	return
}

// this should go here
func (StripeProvider) convertPricingSchedule(p *stripe.Price) PricingSchedule {
	switch p.Type {
	case stripe.PriceTypeOneTime:
		return PricingOnce
	case stripe.PriceTypeRecurring:
		switch p.Recurring.Interval {
		case stripe.PriceRecurringIntervalMonth:
			return PricingMonthly
		case stripe.PriceRecurringIntervalYear:
			return PricingAnnual
		default:
			return ""
		}
	default:
		return ""
	}
}
