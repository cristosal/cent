package serve

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cristosal/cent/client"
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
			client.SubjCheckout:                     srv.handleCheckout(),
			client.SubjCustomerAdd:                  srv.handleAddCustomer(),
			client.SubjCustomerGetByEmail:           srv.handleGetCustomerByEmail(),
			client.SubjCustomerGetByID:              srv.handleGetCustomerByID(),
			client.SubjCustomerGetByProviderID:      srv.handleGetCustomerByProvider(),
			client.SubjCustomerList:                 srv.handleListCustomers(),
			client.SubjCustomerRemoveByProviderID:   srv.handleRemoveCustomerByProviderID(),
			client.SubjPlanAdd:                      srv.handleAddPlan(),
			client.SubjPlanGetByID:                  srv.handleGetPlanByID(),
			client.SubjPlanGetByName:                srv.handleGetPlanByName(),
			client.SubjPlanGetByProviderID:          srv.handleGetPlanByProviderID(),
			client.SubjPlanGetBySubscriptionID:      srv.handleGetPlanBySubscriptionID(),
			client.SubjPlanList:                     srv.handleListPlans(),
			client.SubjPlanListActive:               srv.handleListActivePlans(),
			client.SubjPlanListByUsername:           srv.handleGetPlansByUsername(),
			client.SubjPlanRemoveByProviderID:       srv.handleRemovePlanByProviderID(),
			client.SubjPriceAdd:                     srv.handleAddPrice(),
			client.SubjPriceGetByID:                 srv.handleGetPriceByID(),
			client.SubjPriceGetByProviderID:         srv.handleGetPriceByProviderID(),
			client.SubjPriceList:                    srv.handleListPrices(),
			client.SubjPriceListByPlanID:            srv.handleListPricesByPlanID(),
			client.SubjSubscriptionGetByID:          srv.handleGetSubscriptionByID(),
			client.SubjSubscriptionGetByProviderID:  srv.handleGetSubscriptionByProviderID(),
			client.SubjSubscriptionList:             srv.handleListSubscriptions(),
			client.SubjSubscriptionListByCustomerID: srv.handleListSubscriptionsByCustomerID(),
			client.SubjSubscriptionListByPlanID:     srv.handleListSubscriptionsByPlanID(),
			client.SubjSubscriptionListByUsername:   srv.handleListSubscriptionsByUsername(),
			client.SubjSubscriptionUserAdd:          srv.handleAddSubscriptionUser(),
			client.SubjSubscriptionUserCount:        srv.handleCountSubscriptionUsers(),
			client.SubjSubscriptionUserList:         srv.handleListSubscriptionUsers(),
			client.SubjSubscriptionUserRemove:       srv.handleRemoveSubscriptionUser(),
			client.SubjSync:                         srv.handleSync(),
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
