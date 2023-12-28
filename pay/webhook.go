package pay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Webhook returns the http handler that is responsible for handling any event received from stripe
func (s *StripeProvider) Webhook() http.HandlerFunc {
	const MaxBodyBytes = int64(65536)
	whevents := make(chan stripe.Event)

	go func() {
		for event := range whevents {
			log.Printf("stripe webhook: recieved event: %s", event.Type)
			var err error

			switch event.Type {
			case "product.created":
				err = s.handleProductCreated(event.Data)
			case "product.updated":
				err = s.handleProductUpdated(event.Data)
			case "product.deleted":
				err = s.handleProductDeleted(event.Data)
			case "price.created":
				err = s.handlePriceCreated(event.Data)
			case "price.updated":
				err = s.handlePriceUpdated(event.Data)
			case "price.deleted":
				err = s.handlePriceDeleted(event.Data)
			case "customer.created":
				err = s.handleCustomerCreated(event.Data)
			case "customer.updated":
				err = s.handleCustomerUpdated(event.Data)
			case "customer.deleted":
				err = s.handleCustomerDeleted(event.Data)
			case "customer.subscription.created":
				err = s.handleSubscriptionCreated(event.Data)
			case "customer.subscription.updated":
				err = s.handleSubscriptionUpdated(event.Data)
			case "customer.subscription.deleted":
				err = s.handleSubscriptionDeleted(event.Data)
			}

			if err != nil {
				log.Printf("error handling stripe event %s: %v", event.Type, err)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Pass the request body and Stripe-Signature header to ConstructEvent, along
		// with the webhook signing key.
		sig := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, sig, s.config.WebhookSecret)
		if err != nil {
			log.Printf("Error verifying webhook signature: %v\n", err)
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		if s.hasWebhookEvent(ProviderStripe, event.ID) {
			log.Printf("Already processed event with id %s", event.ID)
			w.WriteHeader(http.StatusOK)
			return
		}

		if err := s.addWebhookEvent(&WebhookEvent{
			Provider:   ProviderStripe,
			ProviderID: event.ID,
			EventType:  event.Type,
			Payload:    event.Data.Raw,
		}); err != nil {
			log.Printf("error while saving stripe event: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		whevents <- event
		w.WriteHeader(http.StatusOK)
	}
}

func (s *StripeProvider) handleSubscriptionCreated(data *stripe.EventData) error {
	var sub stripe.Subscription
	if err := sub.UnmarshalJSON(data.Raw); err != nil {
		return err
	}

	subscr, err := s.convertSubscription(&sub)
	if err != nil {
		return err
	}

	return s.addSubscription(subscr)
}

func (s *StripeProvider) handleSubscriptionUpdated(data *stripe.EventData) error {
	var sub stripe.Subscription
	if err := sub.UnmarshalJSON(data.Raw); err != nil {
		return err
	}

	subscr, err := s.convertSubscription(&sub)
	if err != nil {
		return err
	}

	return s.updateSubscriptionByProvider(subscr)
}

func (s *StripeProvider) handleSubscriptionDeleted(data *stripe.EventData) error {
	var sub stripe.Subscription
	if err := sub.UnmarshalJSON(data.Raw); err != nil {
		return err
	}

	subscr, err := s.convertSubscription(&sub)
	if err != nil {
		return err
	}

	return s.removeSubscriptionByProvider(subscr)
}

func (s *StripeProvider) handleCustomerCreated(data *stripe.EventData) error {
	var c stripe.Customer
	if err := json.Unmarshal(data.Raw, &c); err != nil {
		return err
	}
	return s.addCustomer(s.convertCustomer(&c))
}

func (s *StripeProvider) handleCustomerUpdated(data *stripe.EventData) error {
	var c stripe.Customer
	if err := json.Unmarshal(data.Raw, &c); err != nil {
		return err
	}

	return s.updateCustomerByProvider(s.convertCustomer(&c))
}

func (s *StripeProvider) handleCustomerDeleted(data *stripe.EventData) error {
	var c stripe.Customer
	if err := json.Unmarshal(data.Raw, &c); err != nil {
		return err
	}
	return s.removeCustomerByProvider(ProviderStripe, c.ID)
}

func (s *StripeProvider) handlePriceCreated(data *stripe.EventData) error {
	var p stripe.Price
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}

	pr, err := s.convertPrice(&p)
	if err != nil {
		return err
	}

	return s.addPrice(pr)
}

func (s *StripeProvider) handlePriceUpdated(data *stripe.EventData) error {
	var p stripe.Price
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}

	pr, err := s.convertPrice(&p)
	if err != nil {
		return err
	}

	return s.updatePriceByProvider(pr)
}

func (s *StripeProvider) handlePriceDeleted(data *stripe.EventData) error {
	var p stripe.Price
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}

	return s.removePriceByProvider(&Price{
		Provider:   ProviderStripe,
		ProviderID: p.ID,
	})
}

func (s *StripeProvider) handleProductCreated(data *stripe.EventData) error {
	var p stripe.Product
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}

	return s.addPlan(s.convertProduct(&p))
}

func (s *StripeProvider) handleProductUpdated(data *stripe.EventData) error {
	var p stripe.Product
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}

	return s.updatePlanByProvider(s.convertProduct(&p))
}

func (s *StripeProvider) handleProductDeleted(data *stripe.EventData) error {
	var p stripe.Product
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}
	return s.removePlanByProvider(ProviderStripe, p.ID)
}

func (StripeProvider) convertCustomer(c *stripe.Customer) *Customer {
	return &Customer{
		ProviderID: c.ID,
		Provider:   ProviderStripe,
		Name:       c.Name,
		Email:      c.Email,
	}
}

func (StripeProvider) convertProduct(p *stripe.Product) *Plan {
	return &Plan{
		Name:        p.Name,
		Description: p.Description,
		Provider:    ProviderStripe,
		ProviderID:  p.ID,
		Active:      p.Active,
	}
}

func (s *StripeProvider) convertPrice(p *stripe.Price) (*Price, error) {
	pl, err := s.GetPlanByProviderID(ProviderStripe, p.Product.ID)
	if err != nil {
		return nil, err
	}

	pr := &Price{
		Provider:   ProviderStripe,
		ProviderID: p.ID,
		Amount:     p.UnitAmount,
		Currency:   string(p.Currency),
		Schedule:   s.convertPricingSchedule(p),
		TrialDays:  int(p.Recurring.TrialPeriodDays), // TODO: check if this is actually sent through in the webhook
		PlanID:     pl.ID,
	}

	return pr, nil
}

func (s *StripeProvider) convertSubscription(sub *stripe.Subscription) (*Subscription, error) {
	// ensure that the first item is a subscription
	if sub.Items == nil ||
		len(sub.Items.Data) == 0 ||
		sub.Items.Data[0].Price == nil {
		return nil, errors.New("unable to get price id from subscription")
	}

	priceID := sub.Items.Data[0].Price.ID
	pr, err := s.GetPriceByProvider(ProviderStripe, priceID)
	if err != nil {
		return nil, fmt.Errorf("could not get price %s: %w", priceID, err)
	}

	cust, err := s.GetCustomerByProvider(ProviderStripe, sub.Customer.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get customer with provider_id = %s for subscription %s: %w",
			sub.Customer.ID, sub.ID, err)
	}

	subscr := Subscription{
		Provider:   ProviderStripe,
		ProviderID: sub.ID,
		CustomerID: cust.ID,
		PriceID:    pr.ID,
		Active:     sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing,
		CreatedAt:  time.Unix(sub.Created, 0),
	}

	return &subscr, nil
}
