package client

import (
	"errors"

	"github.com/cristosal/pay"
	"github.com/spf13/cobra"
)

var (
	username string
	planID   int64
	sub      pay.Subscription

	SubscriptionsCmd = &cobra.Command{
		Use:     "subscriptions",
		Aliases: []string{"subs"},
		Short:   "Manage subscriptions",
	}

	listSubscriptionsCmd = &cobra.Command{
		Use:                   "list [--plan-id | --username]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"ls"},
		Short:                 "list subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			var subs []pay.Subscription
			if username != "" {
				subs, err = cl.ListSubscriptionsByUsername(username)
			} else if planID != 0 {
				subs, err = cl.ListSubscriptionsByPlanID(planID)
			} else {
				subs, err = cl.ListSubscriptions()
			}

			if err != nil {
				return err
			}

			return pprint(subs)
		},
	}

	getSubscriptionCmd = &cobra.Command{
		Use:                   "get",
		DisableFlagsInUseLine: true,
		Short:                 "get subscription by property",
		Aliases:               []string{"g"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			var s *pay.Subscription
			if sub.CustomerID != 0 {
				s, err = cl.GetSubscriptionsByCustomerID(sub.CustomerID)
			} else if sub.ProviderID != "" {
				s, err = cl.GetSubscriptionsByProviderID(sub.ProviderID)
			} else {
				return errors.New("one of --customer-id or --provider-id is required")
			}

			if err != nil {
				return err
			}

			return pprint(s)
		},
	}
)

func init() {
	listSubscriptionsCmd.Flags().Int64Var(&planID, "plan-id", 0, "list subscriptions by plan id")
	listSubscriptionsCmd.Flags().StringVar(&username, "username", "", "list subscriptions by username")
	listSubscriptionsCmd.MarkFlagsMutuallyExclusive("plan-id", "username")

	getSubscriptionCmd.Flags().Int64Var(&sub.CustomerID, "customer-id", 0, "get subscription by customer id")
	getSubscriptionCmd.Flags().StringVar(&sub.ProviderID, "provider-id", "", "get subscription by provider id")
	getSubscriptionCmd.MarkFlagsOneRequired("customer-id", "provider-id")
	getSubscriptionCmd.MarkFlagsMutuallyExclusive("customer-id", "provider-id")

	SubscriptionsCmd.AddCommand(listSubscriptionsCmd, getSubscriptionCmd)
}
