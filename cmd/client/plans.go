package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cristosal/pay"
	"github.com/spf13/cobra"
)

var (
	pl     pay.Plan
	active bool

	PlansCmd = &cobra.Command{
		Use:   "plans",
		Short: "Manage plans",
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
			if pl.ProviderID == "" {
				return errors.New("provider id is required")
			}

			cl, err := getClient()
			if err != nil {
				return err
			}

			return cl.RemovePlanByProviderID(pl.ProviderID)
		},
	}

	listPlansCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			var plans []pay.Plan

			if pl.Active {
				plans, err = cl.ListActivePlans()
			} else if username != "" {
				plans, err = cl.ListPlansByUsername(username)
			} else {
				plans, err = cl.ListPlans()
			}

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
			} else if sub.ID != 0 {
				plan, err = cl.GetPlanBySubscriptionID(sub.ID)
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
	addPlanCmd.Flags().BoolVar(&pl.Active, "active", true, "plan is active")
	addPlanCmd.MarkFlagRequired("name")

	removePlanCmd.Flags().StringVar(&pl.ProviderID, "provider-id", "", "remove plan in provider")
	removePlanCmd.MarkFlagRequired("provider-id")

	listPlansCmd.Flags().StringVar(&username, "username", "", "list plans by username")
	listPlansCmd.Flags().BoolVar(&pl.Active, "active", false, "list active plans")
	listPlansCmd.MarkFlagsMutuallyExclusive("username", "active")

	getPlanCmd.Flags().Int64Var(&pl.ID, "id", 0, "get plan by id")
	getPlanCmd.Flags().Int64Var(&sub.ID, "subscription-id", 0, "get plan by subscription id")
	getPlanCmd.Flags().StringVar(&pl.Name, "name", "", "get plan by name")
	getPlanCmd.Flags().StringVar(&pl.ProviderID, "provider-id", "", "get plan by provider id")
	getPlanCmd.MarkFlagsMutuallyExclusive("id", "subscription-id", "name", "provider-id")
	getPlanCmd.MarkFlagsOneRequired("id", "subscription-id", "name", "provider-id")

	PlansCmd.AddCommand(addPlanCmd, removePlanCmd, listPlansCmd, getPlanCmd)
}

func pprint(v any) error {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
