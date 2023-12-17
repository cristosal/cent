package client

import (
	"github.com/cristosal/pay"
	"github.com/spf13/cobra"
)

var (
	pl pay.Plan

	PlansCmd = &cobra.Command{
		Use:   "plans",
		Short: "manage plans",
	}

	addPlanCmd = &cobra.Command{
		Use:   "add",
		Short: "add plan to provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			return cl.AddPlan(&pl)
		},
	}
)

func init() {
	addPlanCmd.Flags().StringVar(&pl.Name, "name", "", "plan name")
	addPlanCmd.Flags().StringVar(&pl.Description, "desc", "", "plan description")
	addPlanCmd.Flags().StringVar(&pl.Provider, "provider", pay.ProviderStripe, "plan provider")
	addPlanCmd.MarkFlagRequired("name")
	PlansCmd.AddCommand(addPlanCmd)
}
