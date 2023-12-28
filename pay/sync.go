package pay

import (
	"errors"
	"fmt"
	"log"

	"github.com/cristosal/orm"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/subscription"
)

// Sync repository data with stripe
func (s *StripeProvider) Sync() error {
	if err := s.syncCustomers(); err != nil {
		return fmt.Errorf("error syncing customers: %w", err)
	}

	if err := s.syncPlans(); err != nil {
		return fmt.Errorf("error syncing plans: %w", err)
	}

	if err := s.syncPrices(); err != nil {
		return fmt.Errorf("error syncing prices: %w", err)
	}

	if err := s.syncSubscriptions(); err != nil {
		return fmt.Errorf("error syncing subscriptions: %w", err)
	}

	return nil
}

func (s *StripeProvider) syncPrices() error {
	it := price.List(nil)
	var ids []string

	for it.Next() {
		p := it.Price()
		ids = append(ids, p.ID)

		pr, err := s.convertPrice(p)
		if err != nil {
			log.Printf("error converting price %s: %v", p.ID, err)
			continue
		}

		_, err = s.GetPriceByProvider(ProviderStripe, p.ID)

		// we did not find the price for the provider
		if errors.Is(err, orm.ErrNotFound) {
			if err := s.addPrice(pr); err != nil {
				log.Printf("error updating price %s: %v", pr.ProviderID, err)
			}
			continue
		}

		if err != nil {
			log.Printf("error getting price %s: %v", p.ID, err)
			continue
		}

		if err := s.updatePriceByProvider(pr); err != nil {
			log.Printf("error adding price %s: %v", pr.ProviderID, err)
			continue
		}
	}

	if err := it.Err(); err != nil {
		return err
	}

	return s.removePriceOrphans(ProviderStripe, ids)
}

func (s *StripeProvider) syncCustomers() error {
	var ids []string

	it := customer.List(nil)
	for it.Next() {
		cust := it.Customer()
		ids = append(ids, cust.ID)
		c := s.convertCustomer(cust)

		found, _ := s.GetCustomerByProvider(ProviderStripe, cust.ID)
		if found == nil {
			if err := s.addCustomer(c); err != nil {
				log.Printf("error while adding stripe customer with id %s: %v", c.ProviderID, err)
			}
			continue
		}

		c.ID = found.ID
		if c.Name != found.Name || c.Email != found.Email {
			if err := s.updateCustomerByProvider(c); err != nil {
				log.Printf("error while updating stripe customer with id %s: %v", c.ProviderID, err)
			}
		}
	}

	if it.Err() != nil {
		return it.Err()
	}

	return s.removeCustomerOrphans(ProviderStripe, ids)
}

func (s *StripeProvider) syncPlans() error {
	it := product.List(nil)
	var ids []string

	for it.Next() {
		p := it.Product()
		ids = append(ids, p.ID)
		pl := s.convertProduct(p)

		// we need to see if we already have it
		_, err := s.GetPlanByProviderID(ProviderStripe, p.ID)

		if errors.Is(err, orm.ErrNotFound) {
			if err := s.addPlan(pl); err != nil {
				log.Printf("error while adding plan %s: %v", p.ID, err)
			}
			continue
		}

		if err != nil {
			log.Printf("error while getting plan %s: %v", p.ID, err)
			continue
		}

		if err := s.updatePlanByProvider(pl); err != nil {
			log.Printf("error updating plan %s: %v", pl.ProviderID, err)
		}
	}

	if err := it.Err(); err != nil {
		return err
	}

	return s.removePlanOrphans(ProviderStripe, ids)
}

// syncSubscriptions pulls in all subscriptions from stripe
func (s *StripeProvider) syncSubscriptions() error {
	var ids []string
	it := subscription.List(nil)
	for it.Next() {
		sub := it.Subscription()
		ids = append(ids, sub.ID)

		subscr, err := s.convertSubscription(sub)
		if err != nil {
			log.Printf("error converting subscription %s: %v", sub.ID, err)
		}

		_, err = s.GetSubscriptionByProvider(ProviderStripe, sub.ID)
		if errors.Is(err, orm.ErrNotFound) {
			// we add it
			if err := s.addSubscription(subscr); err != nil {
				log.Printf("error adding subscription %s: %v", subscr.ProviderID, err)
			}
			continue
		}

		if err != nil {
			log.Printf("error getting subscription %s: %v", sub.ID, err)
			continue
		}

		if err := s.updateSubscriptionByProvider(subscr); err != nil {
			log.Printf("error updating subscription %s: %v", subscr.ProviderID, err)
			continue
		}
	}

	if err := it.Err(); err != nil {
		return err
	}

	return s.removeSubscriptionOrphans(ProviderStripe, ids)
}

func convertStringsToInterfaces(input []string) []interface{} {
	var result []interface{}
	for _, v := range input {
		result = append(result, v)
	}
	return result
}
