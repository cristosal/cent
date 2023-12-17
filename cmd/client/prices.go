package client

import (
	"errors"

	"github.com/cristosal/pay"
	"github.com/spf13/cobra"
)

var (
	pr pay.Price

	PricesCmd = &cobra.Command{
		Use:   "prices",
		Short: "Manage prices",
	}

	addPriceCmd = &cobra.Command{
		Use:   "add",
		Short: "add price to provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}
			return cl.AddPrice(&pr)
		},
	}

	listPricesCmd = &cobra.Command{
		Use:     "list",
		Short:   "list prices",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}
			var prices []pay.Price
			if pr.PlanID != 0 {
				prices, err = cl.ListPricesByPlanID(pr.PlanID)
			} else {
				prices, err = cl.ListPrices()
			}
			if err != nil {
				return err
			}

			return pprint(prices)
		},
	}

	getPriceCmd = &cobra.Command{
		Use:     "get",
		Short:   "get price by property",
		Aliases: []string{"g"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			var price *pay.Price

			if pr.ID != 0 {
				price, err = cl.GetPriceByID(pr.ID)
			} else if pr.ProviderID != "" {
				price, err = cl.GetPriceByProviderID(pr.ProviderID)
			} else {
				return errors.New("must specify one of --id or --provider-id")
			}

			if err != nil {
				return err
			}

			return pprint(price)
		},
	}
)

func init() {
	addPriceCmd.Flags().StringVar(&pr.Currency, "currency", "USD", "price currency")
	addPriceCmd.Flags().Int64Var(&pr.Amount, "amount", 0, "unit amount (cents)")
	addPriceCmd.Flags().StringVar(&pr.Schedule, "schedule", pay.PricingMonthly, "pricing schedule (monthly or annual)")
	addPriceCmd.Flags().Int64Var(&pr.PlanID, "plan-id", 0, "plan to which this price belongs")
	addPriceCmd.Flags().IntVar(&pr.TrialDays, "trial-days", 0, "amount of trial days for plan")
	addPriceCmd.MarkFlagRequired("plan-id")

	listPricesCmd.Flags().Int64Var(&pr.PlanID, "plan-id", 0, "list prices for plan")

	getPriceCmd.Flags().Int64Var(&pr.ID, "id", 0, "price id")
	getPriceCmd.Flags().StringVar(&pr.ProviderID, "provider-id", "", "price provider id")
	getPriceCmd.MarkFlagsMutuallyExclusive("id", "provider-id")
	getPriceCmd.MarkFlagsOneRequired("id", "provider-id")

	PricesCmd.AddCommand(addPriceCmd, listPricesCmd, getPriceCmd)
}
