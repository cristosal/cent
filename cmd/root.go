package cmd

import (
	"github.com/cristosal/cent/cmd/client"
	"github.com/cristosal/cent/cmd/serve"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "cent",
	Short: "CLI for managing cent microservice",
}

func init() {
	Root.AddCommand(serve.Cmd, client.CustomersCmd, client.PlansCmd, client.PricesCmd, client.SubscriptionsCmd)
}
