package serve

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

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

var Cmd = &cobra.Command{
	Use:   "serve",
	Short: "serves cent microservice",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := sql.Open("pgx", pgConnectionString)
		if err != nil {
			return fmt.Errorf("error opening database: %w", err)
		}

		p := pay.NewStripeProvider(&pay.StripeConfig{
			Repo:          pay.NewEntityRepo(db),
			Key:           stripeApiKey,
			WebhookSecret: stripeWebhookSecret,
		})

		log.Println("syncing...")
		if err := p.Sync(); err != nil {
			log.Fatal(fmt.Errorf("sync error: %w", err))
		}
		log.Println("done")

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
			"cent.customer.add":                 srv.handleAddCustomer(),
			"cent.customer.get.email":           srv.handleGetCustomerByEmail(),
			"cent.customer.get.id":              srv.handleGetCustomerByID(),
			"cent.customer.get.provider_id":     srv.handleGetCustomerByProvider(),
			"cent.customer.list":                srv.handleListCustomers(),
			"cent.customer.remove":              srv.handleRemoveCustomer(),
			"cent.plan.add":                     srv.handleAddPlan(),
			"cent.plan.get.id":                  srv.handleGetPlanByID(),
			"cent.plan.get.subscription_id":     srv.handleGetPlanBySubscriptionID(),
			"cent.plan.get.name":                srv.handleGetPlanByName(),
			"cent.plan.get.provider_id":         srv.handleGetPlanByProvider(),
			"cent.plan.list":                    srv.handleListPlans(),
			"cent.plan.list.username":           srv.handleGetPlansByUsername(),
			"cent.plan.remove":                  srv.handleRemovePlan(),
			"cent.price.add":                    srv.handleAddPrice(),
			"cent.price.list":                   srv.handleListPrices(),
			"cent.price.list.plan_id":           srv.handleListPricesByPlanID(),
			"cent.price.get.id":                 srv.handleGetPriceByID(),
			"cent.price.get.provider_id":        srv.handleGetPriceByProviderID(),
			"cent.subscription.list":            srv.handleListSubscriptions(),
			"cent.subscription.list.username":   srv.handleListSubscriptionsByUsername(),
			"cent.subscription.list.plan_id":    srv.handleListSubscriptionsByPlanID(),
			"cent.subscription.get.customer_id": srv.handleGetSubscriptionByCustomerID(),
			"cent.subscription.get.provider_id": srv.handleGetSubscriptionByProviderID(),
			"cent.subscription.user.add":        srv.handleAddSubscriptionUser(),
			"cent.subscription.user.count":      srv.handleCountSubscriptionUsers(),
			"cent.subscription.user.list":       srv.handleListSubscriptionUsers(),
			"cent.subscription.user.remove":     srv.handleRemoveSubscriptionUser(),
			"cent.sync":                         srv.handleSync(),
			"cent.checkout":                     srv.handleCheckout(),
		}

		var subs []*nats.Subscription
		for k, v := range submap {
			sub, err := srv.sub(k, v)
			if err != nil {
				return fmt.Errorf("error subscribing to %s: %w", k, err)
			}
			subs = append(subs, sub)
		}

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		fmt.Println("listening...")

		http.ListenAndServe(addr, nil)

		<-ch
		for _, sub := range subs {
			sub.Unsubscribe()
		}

		return nil
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&natsURL, "nats", "", nats.DefaultURL, "NATS connection url")
	Cmd.PersistentFlags().StringVarP(&natsURL, "pg", "", "", "Postgres connection string")
	Cmd.PersistentFlags().StringVarP(&stripeApiKey, "stripe-api-key", "", "", "Stripe api key from stripe account")
	Cmd.PersistentFlags().StringVarP(&stripeWebhookSecret, "stripe-webhook-secret", "", "", "Stripe webhook secret for verifying webhook post requests")
	Cmd.PersistentFlags().StringVarP(&addr, "addr", "", "127:8080", "HTTP server address")
}
