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
	sqlDSN              string
	sqlDriver           string
	stripeApiKey        string
	stripeWebhookSecret string
	enableWebUI         bool
	cmd                 = &cobra.Command{
		Use:   "centd",
		Short: "payment microservice",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := sql.Open(sqlDriver, getConnectionString())
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

			s := cent.New(&cent.Config{
				NatsURL:         natsURL,
				Provider:        p,
				Queue:           "cent",
				HttpAddr:        addr,
				WebhookEndpoint: "/webhook",
				EnableWebUI:     enableWebUI,
			})

			return s.Listen()
		},
	}
)

func init() {
	cmd.Flags().BoolVar(&enableWebUI, "web-ui", false, "Enables Web UI")
	cmd.Flags().StringVar(&natsURL, "nats", nats.DefaultURL, "NATS connection url")
	cmd.Flags().StringVar(&sqlDriver, "sql-driver", "pgx", "SQL Data Source Name")
	cmd.Flags().StringVar(&sqlDSN, "sql-dsn", "", "SQL Data Source Name")
	cmd.Flags().StringVar(&stripeApiKey, "stripe-api-key", "", "Stripe api key from stripe account")
	cmd.Flags().StringVar(&stripeWebhookSecret, "stripe-webhook-secret", "", "Stripe webhook secret for verifying webhook post requests")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8080", "HTTP server address")
}

func getConnectionString() string {
	if sqlDSN == "" {
		return os.Getenv("CONNECTION_STRING")
	}
	return sqlDSN
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
