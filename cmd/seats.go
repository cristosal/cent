package cmd

import (
	"fmt"
	"strconv"

	"github.com/cristosal/pay"
	"github.com/spf13/cobra"
)

var (
	su pay.SubscriptionUser

	SeatsCmd = &cobra.Command{
		Use:   "seats",
		Short: "Manage seats for subscriptions",
	}

	addSeatCmd = &cobra.Command{
		Use:   "add",
		Short: "add a seat to a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			return cl.AddSubscriptionUser(&su)
		},
	}

	removeSeatCmd = &cobra.Command{
		Use:     "remove",
		Short:   "remove a seat from a subscription",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			return cl.RemoveSubscriptionUser(&su)
		},
	}

	countSeatCmd = &cobra.Command{
		Use:   "count",
		Short: "count seats for a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			n, err := cl.CountSubscriptionUsers(su.SubscriptionID)
			if err != nil {
				return err
			}

			fmt.Println(strconv.FormatInt(n, 10))
			return nil
		},
	}

	listSeatCmd = &cobra.Command{
		Use:     "list",
		Short:   "list seats in a subscription",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getClient()
			if err != nil {
				return err
			}

			usernames, err := cl.ListSubscriptionUsers(su.SubscriptionID)
			if err != nil {
				return err
			}

			return pprint(usernames)
		},
	}
)

func init() {
	addSeatCmd.Flags().Int64Var(&su.SubscriptionID, "subscription-id", 0, "subscription id")
	addSeatCmd.Flags().StringVar(&su.Username, "username", "", "username")
	addSeatCmd.MarkFlagRequired("subscription-id")
	addSeatCmd.MarkFlagRequired("username")

	removeSeatCmd.Flags().Int64Var(&su.SubscriptionID, "subscription-id", 0, "subscription id")
	removeSeatCmd.Flags().StringVar(&su.Username, "username", "", "username")
	removeSeatCmd.MarkFlagRequired("subscription-id")
	removeSeatCmd.MarkFlagRequired("username")

	countSeatCmd.Flags().Int64Var(&su.SubscriptionID, "subscription-id", 0, "subscription id")
	countSeatCmd.MarkFlagRequired("subscription-id")

	listSeatCmd.Flags().Int64Var(&su.SubscriptionID, "subscription-id", 0, "subscription id")
	listSeatCmd.MarkFlagRequired("subscription-id")

	SeatsCmd.AddCommand(addSeatCmd, removeSeatCmd, countSeatCmd, listSeatCmd)
}
