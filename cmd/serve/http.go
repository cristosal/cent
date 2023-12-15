package serve

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/cristosal/cent/templates"
	"github.com/cristosal/orm"

	"github.com/cristosal/pay"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func handleFuncs(p *pay.StripeProvider) {
	http.HandleFunc("/", handleHome())
	http.HandleFunc("/webhook", p.Webhook())
	http.HandleFunc("/plans", handlePlans(p))
	http.HandleFunc("/plans/new", handlePlansNew(p))
	http.HandleFunc("/plans/delete", handlePlansDelete(p))
	http.HandleFunc("/prices/", handlePrices(p))
	http.HandleFunc("/prices/new", handlePricesNew(p))
	http.HandleFunc("/sync", handleSync(p))
	http.HandleFunc("/customers", handleCustomers(p))
	http.HandleFunc("/customers/new", handleCustomersNew(p))
	http.HandleFunc("/subscriptions", handleSubscriptions(p))
	http.HandleFunc("/subscriptions/users", handleSubscriptionsUsers(p))
	http.HandleFunc("/subscriptions/users/new", handleSubscriptionsUsersNew(p))
	http.HandleFunc("/subscriptions/users/delete", handleSubscriptionsUsersDelete(p))
	http.HandleFunc("/events", handleWebhookEvents(p))
	http.HandleFunc("/checkout", handleCheckout(p))
	http.HandleFunc("/checkout/success", handleCheckoutSuccess())
}

func handleSubscriptionsUsersDelete(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		var (
			q        = r.URL.Query()
			username = q.Get("username")
			s        = q.Get("s")
		)

		subID, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		if err := p.RemoveSubscriptionUser(&pay.SubscriptionUser{
			SubscriptionID: subID,
			Username:       username,
		}); err != nil {
			return err
		}

		redirect := fmt.Sprintf("/subscriptions/users?s=%s", s)
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return nil
	})
}

func handleSubscriptionsUsersNew(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				return err
			}

			username := strings.Trim(r.FormValue("username"), " ")
			subIDString := r.URL.Query().Get("s")

			subID, err := strconv.ParseInt(subIDString, 10, 64)
			if err != nil {
				return err
			}

			if err := p.AddSubscriptionUser(&pay.SubscriptionUser{
				SubscriptionID: subID,
				Username:       username,
			}); err != nil {
				return err
			}

			http.Redirect(w, r, fmt.Sprintf("/subscriptions/users?s=%d", subID), http.StatusSeeOther)
			return nil
		}

		return errors.New("method not supported")
	})
}

func handleSubscriptionsUsers(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		subIDString := r.URL.Query().Get("s")
		if subIDString == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return nil
		}

		subID, err := strconv.ParseInt(subIDString, 10, 64)
		if err != nil {
			return err
		}

		usernames, err := p.ListUsernames(subID)
		if errors.Is(err, orm.ErrNotFound) {
			usernames = []string{}
			err = nil
		}

		if err != nil {
			return err
		}

		return templates.SubscriptionUsers(subID, usernames).Render(r.Context(), w)
	})
}

func handleCheckoutSuccess() http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		return templates.CheckoutSuccess().Render(r.Context(), w)
	})
}

func handleCheckout(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodPost:
			var (
				formCustomerID = r.FormValue("customer_id")
				formPriceID    = r.FormValue("customer_id")
			)

			customerID, err := strconv.ParseInt(formCustomerID, 10, 64)
			if err != nil {
				return err
			}

			priceID, err := strconv.ParseInt(formPriceID, 10, 64)
			if err != nil {
				return err
			}

			url, err := p.Checkout(&pay.CheckoutRequest{
				CustomerID:  customerID,
				PriceID:     priceID,
				RedirectURL: "http://" + addr + "/checkout/success",
			})
			if err != nil {
				return err
			}

			http.Redirect(w, r, url, http.StatusSeeOther)
		default:
			customers, err := p.ListAllCustomers()
			if err != nil {
				return err
			}

			prices, err := p.ListAllPrices()
			if err != nil {
				return err
			}

			return templates.CheckoutForm(customers, prices).Render(r.Context(), w)
		}
		return nil
	})
}

func handleSubscriptions(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		username := r.URL.Query().Get("username")
		var (
			subs []pay.Subscription
			err  error
		)

		if username == "" {
			subs, err = p.ListAllSubscriptions()
			if errors.Is(err, orm.ErrNotFound) {
				err = nil
			}

			if err != nil {
				return err
			}
		} else {
			subs, err = p.ListSubscriptionsByUsername(username)
			if errors.Is(err, pay.ErrSubscriptionNotFound) {
				err = nil
			}

			if err != nil {
				return err
			}
		}

		return templates.SubscriptionsIndex(subs, username).Render(r.Context(), w)
	})
}

func handleWebhookEvents(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		events, err := p.ListAllWebhookEvents()
		if err != nil {
			return err
		}
		return templates.WebhookIndex(events).Render(r.Context(), w)
	})
}

func handleSync(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		if err := p.Sync(); err != nil {
			return err
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	})
}

func handleHome() http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		return templates.Home().Render(r.Context(), w)
	})
}

func handlePlansNew(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				return fmt.Errorf("error while parsing form: %v", err)
			}

			name := r.Form.Get("name")
			desc := r.Form.Get("description")
			active := r.Form.Get("active") == "on"

			err := p.AddPlan(&pay.Plan{
				Name:        name,
				Description: desc,
				Active:      active,
			})
			if err != nil {
				return fmt.Errorf("error while adding plan: %v", err)
			}

			http.Redirect(w, r, "/plans", http.StatusSeeOther)
			return nil
		}

		return templates.PlansNew().Render(r.Context(), w)
	})
}

func handlePlans(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				return err
			}

			name := r.Form.Get("name")
			desc := r.Form.Get("description")
			active := r.Form.Get("active") == "on"
			err := p.AddPlan(&pay.Plan{
				Name:        name,
				Description: desc,
				Active:      active,
			})
			if err != nil {
				return err
			}

			http.Redirect(w, r, "/plans", http.StatusSeeOther)
			return nil
		}

		plans, err := p.ListActivePlans()
		if err != nil {
			return err
		}

		return templates.PlansIndex(plans).Render(r.Context(), w)
	})
}

func handlePlansDelete(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		idQuery := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idQuery, 10, 64)
		if err != nil {
			return err
		}

		pl, err := p.GetPlanByID(id)
		if err != nil {
			return err
		}

		if err := p.RemovePlan(pl); err != nil {
			return err
		}

		http.Redirect(w, r, "/plans", http.StatusSeeOther)
		return nil
	})
}

func handlePrices(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		prices, err := p.ListAllPrices()
		if err != nil {
			return err
		}

		return templates.PricesIndex(prices).Render(r.Context(), w)
	})
}

func handlePricesNew(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				return err
			}

			var (
				planID    = r.FormValue("plan_id")
				amount    = r.FormValue("amount")
				currency  = r.FormValue("currency")
				schedule  = r.FormValue("schedule")
				trialDays = r.FormValue("trial_days")
			)

			parsedAmount, err := strconv.Atoi(amount)
			if err != nil {
				return fmt.Errorf("error parsing amount: %w", err)
			}

			parsedPlanID, err := strconv.Atoi(planID)
			if err != nil {
				return fmt.Errorf("error parsing plan id: %w", err)
			}

			parsedTrialDays, err := strconv.Atoi(trialDays)
			if err != nil {
				return fmt.Errorf("error parsing trial days: %w", err)
			}

			pr := pay.Price{
				Amount:    int64(parsedAmount),
				PlanID:    int64(parsedPlanID),
				Currency:  currency,
				Schedule:  schedule,
				TrialDays: parsedTrialDays,
			}

			if err := p.AddPrice(&pr); err != nil {
				return fmt.Errorf("error adding price: %w", err)
			}

			http.Redirect(w, r, "/prices", http.StatusSeeOther)
			return nil
		default:
			plans, err := p.ListActivePlans()
			if err != nil {
				return err
			}

			return templates.PricesNew(
				pay.PricingMonthly,
				pay.PricingAnnual,
				plans,
			).Render(r.Context(), w)

		}
	})
}

func handleCustomersNew(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				return err
			}

			var (
				name  = r.FormValue("name")
				email = r.FormValue("email")
			)

			if err := p.AddCustomer(&pay.Customer{
				Name:  name,
				Email: email,
			}); err != nil {
				return err
			}

			http.Redirect(w, r, "/customers", http.StatusSeeOther)
		default:
			return templates.CustomersNew().Render(r.Context(), w)
		}
		return nil
	})
}

func handleCustomers(p *pay.StripeProvider) http.HandlerFunc {
	return wrap(func(w http.ResponseWriter, r *http.Request) error {
		customers, err := p.ListAllCustomers()
		if err != nil {
			return err
		}

		return templates.CustomersIndex(customers).Render(r.Context(), w)
	})
}

func writeErr(w http.ResponseWriter, err error) {
	log.Printf("ERROR: %v", err)
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

type wrappedHandlerFunc func(w http.ResponseWriter, r *http.Request) error

func wrap(h wrappedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			writeErr(w, err)
		}
	}
}
