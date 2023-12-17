package client

import (
	"encoding/json"
	"errors"
	"fmt"

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

	removePlanCmd = &cobra.Command{
		Use:     "remove",
		Short:   "remove plan by provider id",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("provider id is required")
			}

			cl, err := getClient()
			if err != nil {
				return err
			}

			for _, id := range args {
				if err := cl.RemovePlanByProviderID(id); err != nil {
					return fmt.Errorf("error removing plan with id %s: %w", id, err)
				}

			}

			return nil
		},
	}

	listPlans = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			plans, err := cl.ListPlans()
			if err != nil {
				return err
			}

			return pprint(plans)
		},
	}

	getPlanCmd = &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "get plan by property",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				plan *pay.Plan
				err  error
			)

			cl, err := getClient()
			if err != nil {
				return err
			}

			if pl.ID != 0 {
				plan, err = cl.GetPlanByID(pl.ID)
			} else if pl.ProviderID != "" {
				plan, err = cl.GetPlanByProviderID(pl.ProviderID)
			} else if pl.Name != "" {
				plan, err = cl.GetPlanByName(pl.Name)
			} else {
				return errors.New("must specify one of --id, --provider-id or --email flags")
			}

			if err != nil {
				return err
			}

			return pprint(plan)
		},
	}
)

func init() {
	addPlanCmd.Flags().StringVar(&pl.Name, "name", "", "plan name")
	addPlanCmd.Flags().StringVar(&pl.Description, "desc", "", "plan description")
	addPlanCmd.Flags().StringVar(&pl.Provider, "provider", pay.ProviderStripe, "plan provider")
	addPlanCmd.MarkFlagRequired("name")
	getPlanCmd.Flags().Int64Var(&pl.ID, "id", 0, "plan id")
	getPlanCmd.Flags().StringVar(&pl.Name, "name", "", "plan name")
	getPlanCmd.Flags().StringVar(&pl.ProviderID, "provider-id", "", "plan provider id")
	PlansCmd.AddCommand(addPlanCmd, removePlanCmd, listPlans, getPlanCmd)
}

func pprint(v any) error {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
