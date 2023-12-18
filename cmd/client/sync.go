package client

import (
	"fmt"

	"github.com/cristosal/pay"
	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync entities",
	RunE: func(cmd *cobra.Command, args []string) error {
		cl, err := getClient()
		if err != nil {
			return err
		}

		return cl.Sync()
	},
}

var checkoutRequest pay.CheckoutRequest

var CheckoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "get checkout url",
	RunE: func(cmd *cobra.Command, args []string) error {
		cl, err := getClient()
		if err != nil {
			return nil
		}

		url, err := cl.Checkout(&checkoutRequest)
		if err != nil {
			return nil
		}

		fmt.Println(url)
		return nil
	},
}

func init() {
	CheckoutCmd.Flags().Int64Var(&checkoutRequest.CustomerID, "customer-id", 0, "customer id")
	CheckoutCmd.Flags().Int64Var(&checkoutRequest.PriceID, "price-id", 0, "price id")
	CheckoutCmd.Flags().StringVar(&checkoutRequest.RedirectURL, "redirect-url", "", "redirect-url")
	CheckoutCmd.MarkFlagRequired("customer-id")
	CheckoutCmd.MarkFlagRequired("price-id")
}
