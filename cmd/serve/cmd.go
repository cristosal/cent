package serve

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cristosal/pay"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

var (
	natsURL             string
	addr                string
	pgConnectionString  string
	stripeApiKey        string
	stripeWebhookSecret string
)

func getConnectionString() string {
	if pgConnectionString == "" {
		return os.Getenv("CONNECTION_STRING")
	}
	return pgConnectionString
}

func getStripeWebhookSecret() string {
	if stripeWebhookSecret == "" {
		return os.Getenv("STRIPE_WEBHOOK_SECRET")
	}

	return stripeWebhookSecret
}

func getStripeApiKey() string {
	if stripeApiKey == "" {
		return os.Getenv("STRIPE_API_KEY")
	}

	return stripeApiKey
}

var Cmd = &cobra.Command{
	Use:   "serve",
	Short: "serves cent microservice",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := sql.Open("pgx", getConnectionString())
		if err != nil {
			return fmt.Errorf("error opening database: %w", err)
		}

		p := pay.NewStripeProvider(&pay.StripeConfig{
			Repo:          pay.NewEntityRepo(db),
			Key:           getStripeApiKey(),
			WebhookSecret: getStripeWebhookSecret(),
		})

		if err := p.Init(); err != nil {
			return fmt.Errorf("error initializing pay: %w", err)
		}

		fmt.Println("syncing...")
		if err := p.Sync(); err != nil {
			log.Fatal(fmt.Errorf("sync error: %w", err))
		}

		handleFuncs(p)
		srv, err := newNatsServer(&natsServerConfig{
			NatsURL:  natsURL,
			Queue:    "cent",
			Provider: p,
		})
		if err != nil {
			return fmt.Errorf("error initializing nats server: %w", err)
		}

		submap := map[string]natsHandler{
			// customer
			"cent.customer.add":                srv.handleAddCustomer(),
			"cent.customer.get.id":             srv.handleGetCustomerByID(),
			"cent.customer.get.email":          srv.handleGetCustomerByEmail(),
			"cent.customer.get.provider_id":    srv.handleGetCustomerByProvider(),
			"cent.customer.list":               srv.handleListCustomers(),
			"cent.customer.remove.provider_id": srv.handleRemoveCustomerByProviderID(),
			// plan
			"cent.plan.add":                 srv.handleAddPlan(),
			"cent.plan.get.id":              srv.handleGetPlanByID(),
			"cent.plan.get.subscription_id": srv.handleGetPlanBySubscriptionID(),
			"cent.plan.get.name":            srv.handleGetPlanByName(),
			"cent.plan.get.provider_id":     srv.handleGetPlanByProviderID(),
			"cent.plan.list":                srv.handleListPlans(),
			"cent.plan.list.active":         srv.handleListActivePlans(),
			"cent.plan.list.username":       srv.handleGetPlansByUsername(),
			"cent.plan.remove.provider_id":  srv.handleRemovePlanByProviderID(),
			// price
			"cent.price.add":             srv.handleAddPrice(),
			"cent.price.list":            srv.handleListPrices(),
			"cent.price.list.plan_id":    srv.handleListPricesByPlanID(),
			"cent.price.get.id":          srv.handleGetPriceByID(),
			"cent.price.get.provider_id": srv.handleGetPriceByProviderID(),
			// subscription
			"cent.subscription.list":             srv.handleListSubscriptions(),
			"cent.subscription.list.username":    srv.handleListSubscriptionsByUsername(),
			"cent.subscription.list.plan_id":     srv.handleListSubscriptionsByPlanID(),
			"cent.subscription.list.customer_id": srv.handleListSubscriptionsByCustomerID(),
			"cent.subscription.get.id":           srv.handleGetSubscriptionByID(),
			"cent.subscription.get.provider_id":  srv.handleGetSubscriptionByProviderID(),
			// user
			"cent.subscription.user.add":    srv.handleAddSubscriptionUser(),
			"cent.subscription.user.count":  srv.handleCountSubscriptionUsers(),
			"cent.subscription.user.list":   srv.handleListSubscriptionUsers(),
			"cent.subscription.user.remove": srv.handleRemoveSubscriptionUser(),
			// other
			"cent.sync":     srv.handleSync(),
			"cent.checkout": srv.handleCheckout(),
		}

		var subs []*nats.Subscription
		for k, v := range submap {
			sub, err := srv.sub(k, v)
			if err != nil {
				return fmt.Errorf("error subscribing to %s: %w", k, err)
			}
			subs = append(subs, sub)
		}

		fmt.Printf("listening on %s...\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			for _, sub := range subs {
				sub.Unsubscribe()
			}
		}

		return nil
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&natsURL, "nats", "", nats.DefaultURL, "NATS connection url")
	Cmd.PersistentFlags().StringVarP(&natsURL, "pg", "", "", "Postgres connection string")
	Cmd.PersistentFlags().StringVarP(&stripeApiKey, "stripe-api-key", "", "", "Stripe api key from stripe account")
	Cmd.PersistentFlags().StringVarP(&stripeWebhookSecret, "stripe-webhook-secret", "", "", "Stripe webhook secret for verifying webhook post requests")
	Cmd.PersistentFlags().StringVarP(&addr, "addr", "", "127.0.0.1:8080", "HTTP server address")
}
