package cmd

import (
	"github.com/cristosal/cent/cmd/serve"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "cent",
	Short: "CLI for managing cent microservice",
}

func init() {
	Root.AddCommand(
		serve.Cmd,
		CustomersCmd,
		PlansCmd,
		PricesCmd,
		SubscriptionsCmd,
		SeatsCmd,
		SyncCmd,
		CheckoutCmd,
	)
}
