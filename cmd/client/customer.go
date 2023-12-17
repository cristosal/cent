package client

import (
	"errors"
	"fmt"
	"time"

	cl "github.com/cristosal/cent/client"
	"github.com/cristosal/pay"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

var (
	c             pay.Customer
	natsURL       string
	clientTimeout time.Duration
)

var (
	CustomersCmd = &cobra.Command{
		Use:          "customers",
		Short:        "manage customers",
		SilenceUsage: true,
	}

	ListCustomersCmd = &cobra.Command{
		Use:   "list",
		Short: "lists all customers",
		RunE: func(cmd *cobra.Command, args []string) error {
			nc, err := nats.Connect(natsURL)
			if err != nil {
				return fmt.Errorf("error connecting to nats: %w", err)
			}
			client := cl.NewClient(nc, clientTimeout)
			customers, err := client.ListCustomers()
			if err != nil {
				return err
			}
			return pprint(customers)
		},
	}

	AddCustomerCmd = &cobra.Command{
		Use:   "add",
		Short: "adds a customer directly to the provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			nc, err := nats.Connect(natsURL)
			if err != nil {
				return fmt.Errorf("error connecting to nats: %w", err)
			}
			client := cl.NewClient(nc, clientTimeout)
			return client.AddCustomer(&c)
		},
	}

	GetCustomerCmd = &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "get customer by property",
		RunE: func(cmd *cobra.Command, args []string) error {
			nc, err := nats.Connect(natsURL)
			if err != nil {
				return fmt.Errorf("error connecting to nats: %w", err)
			}

			client := cl.NewClient(nc, clientTimeout)
			var cust *pay.Customer

			if c.ID != 0 {
				cust, err = client.GetCustomerByID(c.ID)
			} else if c.ProviderID != "" {
				cust, err = client.GetCustomerByProviderID(c.ProviderID)
			} else if c.Email != "" {
				cust, err = client.GetCustomerByEmail(c.Email)
			} else {
				return errors.New("please specify one of id, provider-id or email flags")
			}

			if err != nil {
				return err
			}

			return pprint(cust)
		},
	}

	RemoveCustomerCmd = &cobra.Command{
		Use:     "remove provider_id",
		Aliases: []string{"rm"},
		Short:   "removes a customer from the provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("provider_id required")
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			for _, id := range args {
				if err := client.RemoveCustomerByProviderID(id); err != nil {
					return fmt.Errorf("error removing customer %s: %w", id, err)
				}
			}

			return nil
		},
	}
)

func init() {
	addCustomerFlags(AddCustomerCmd)
	getCustomerFlags(GetCustomerCmd)

	CustomersCmd.PersistentFlags().DurationVar(&clientTimeout, "timeout", time.Second*10, "timeout for request")
	CustomersCmd.PersistentFlags().StringVar(&natsURL, "nats", nats.DefaultURL, "nats connection url")
	CustomersCmd.AddCommand(AddCustomerCmd, ListCustomersCmd, RemoveCustomerCmd, GetCustomerCmd)
}

func getCustomerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Int64Var(&c.ID, "id", 0, "customer id")
	cmd.PersistentFlags().StringVarP(&c.Email, "email", "", "", "customer email")
	cmd.PersistentFlags().StringVarP(&c.ProviderID, "provider-id", "", "", "customer provider id")
}

func addCustomerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&c.Name, "name", "", "", "customer name")
	cmd.PersistentFlags().StringVarP(&c.Email, "email", "", "", "customer email")
	cmd.PersistentFlags().StringVarP(&c.ProviderID, "provider-id", "", "", "customer provider id")
	cmd.PersistentFlags().StringVarP(&c.Provider, "provider", "", pay.ProviderStripe, "customer provider")
}

func getClient() (*cl.Client, error) {

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to nats: %w", err)
	}
	client := cl.NewClient(nc, clientTimeout)
	return client, nil
}
