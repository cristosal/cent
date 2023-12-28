package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/cristosal/cent"
	"github.com/cristosal/cent/pay"
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

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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

var cmd = &cobra.Command{
	Use:   "centd",
	Short: "payment microservice",
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

		srv := cent.New(&cent.Config{
			NatsURL:         natsURL,
			Provider:        p,
			Queue:           "cent",
			HttpAddr:        addr,
			WebhookEndpoint: "/webhook",
		})

		return srv.Listen()
	},
}

func init() {
	cmd.PersistentFlags().StringVarP(&natsURL, "nats", "", nats.DefaultURL, "NATS connection url")
	cmd.PersistentFlags().StringVarP(&pgConnectionString, "pg", "", "", "Postgres connection string")
	cmd.PersistentFlags().StringVarP(&stripeApiKey, "stripe-api-key", "", "", "Stripe api key from stripe account")
	cmd.PersistentFlags().StringVarP(&stripeWebhookSecret, "stripe-webhook-secret", "", "", "Stripe webhook secret for verifying webhook post requests")
	cmd.PersistentFlags().StringVarP(&addr, "addr", "", "127.0.0.1:8080", "HTTP server address")
}
