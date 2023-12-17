package client

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/cristosal/pay"
	"github.com/nats-io/nats.go"
)

type Client struct {
	nc *nats.Conn
	to time.Duration
}

func NewClient(nc *nats.Conn, timeout time.Duration) *Client {
	return &Client{
		nc: nc,
		to: timeout,
	}
}

func (c *Client) AddCustomer(cust *pay.Customer) error {
	data, err := json.Marshal(cust)
	if err != nil {
		return err
	}
	_, err = c.request("cent.customer.add", data)
	return err
}

func (c *Client) GetCustomerByEmail(email string) (*pay.Customer, error) {
	data, err := c.request("cent.customer.get.email", []byte(email))
	if err != nil {
		return nil, err
	}

	var cust pay.Customer
	if err := json.Unmarshal(data, &cust); err != nil {
		return nil, err
	}

	return &cust, nil
}

func (c *Client) GetCustomerByID(id int64) (*pay.Customer, error) {
	str := strconv.FormatInt(id, 10)
	data, err := c.request("cent.customer.get.id", []byte(str))
	if err != nil {
		return nil, err
	}

	var cust pay.Customer
	if err := json.Unmarshal(data, &cust); err != nil {
		return nil, err
	}

	return &cust, nil
}

func (c *Client) GetCustomerByProviderID(providerID string) (*pay.Customer, error) {
	data, err := c.request("cent.customer.get.provider_id", []byte(providerID))
	if err != nil {
		return nil, err
	}

	var cust pay.Customer
	if err := json.Unmarshal(data, &cust); err != nil {
		return nil, err
	}

	return &cust, nil
}

func (c *Client) ListCustomers() ([]pay.Customer, error) {
	data, err := c.request("cent.customer.list", nil)
	if err != nil {
		return nil, err
	}

	var customers []pay.Customer
	if err := json.Unmarshal(data, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (c *Client) RemoveCustomerByProviderID(providerID string) error {
	_, err := c.request("cent.customer.remove.provider_id", []byte(providerID))
	return err
}

func (c *Client) AddPlan(p *pay.Plan) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = c.request("cent.plan.add", data)
	return err
}

func (c *Client) RemovePlan(p *pay.Plan) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = c.request("cent.plan.remove", data)
	return err
}

func (c *Client) GetPlanByID(id int64) (*pay.Plan, error) {
	str := strconv.FormatInt(id, 10)
	data, err := c.request("cent.plan.get.id", []byte(str))
	if err != nil {
		return nil, err
	}
	var p pay.Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) GetPlanBySubscriptionID(id int64) (*pay.Plan, error) {
	str := strconv.FormatInt(id, 10)
	data, err := c.request("cent.plan.get.subscription_id", []byte(str))
	if err != nil {
		return nil, err
	}
	var p pay.Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) GetPlanByName(name string) (*pay.Plan, error) {
	data, err := c.request("cent.plan.get.name", []byte(name))
	if err != nil {
		return nil, err
	}
	var p pay.Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) GetPlanByProviderID(providerID string) (*pay.Plan, error) {
	data, err := c.request("cent.plan.get.provider_id", []byte(providerID))
	if err != nil {
		return nil, err
	}
	var p pay.Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) ListPlans() ([]pay.Plan, error) {
	data, err := c.request("cent.plan.list", nil)
	if err != nil {
		return nil, err
	}
	var plans []pay.Plan
	if err := json.Unmarshal(data, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func (c *Client) ListPlansByUsername(username string) ([]pay.Plan, error) {
	data, err := c.request("cent.plan.list.username", []byte(username))
	if err != nil {
		return nil, err
	}
	var plans []pay.Plan
	if err := json.Unmarshal(data, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func (c *Client) AddPrice(p *pay.Price) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = c.request("cent.price.add", data)
	return err
}

func (c *Client) ListPrices() ([]pay.Price, error) {
	data, err := c.request("cent.price.list", nil)
	if err != nil {
		return nil, err
	}
	var prices []pay.Price
	if err := json.Unmarshal(data, &prices); err != nil {
		return nil, err
	}

	return prices, nil
}

func (c *Client) ListPricesByPlanID(id int64) ([]pay.Price, error) {
	str := strconv.FormatInt(id, 10)
	data, err := c.request("cent.price.list.plan_id", []byte(str))
	if err != nil {
		return nil, err
	}
	var prices []pay.Price
	if err := json.Unmarshal(data, &prices); err != nil {
		return nil, err
	}

	return prices, nil
}

func (c *Client) GetPriceByID(id int64) (*pay.Price, error) {
	str := strconv.FormatInt(id, 10)
	data, err := c.request("cent.price.get.id", []byte(str))
	if err != nil {
		return nil, err
	}
	var p pay.Price
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

func (c *Client) GetPriceByProviderID(id string) (*pay.Price, error) {
	data, err := c.request("cent.price.get.provider_id", []byte(id))
	if err != nil {
		return nil, err
	}
	var p pay.Price
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

func (c *Client) request(subj string, req []byte) ([]byte, error) {
	msg, err := c.nc.Request(subj, req, c.to)
	if err != nil {
		return nil, err
	}

	var resp response
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return nil, err

	}

	if !resp.Success {
		return nil, errors.New(resp.Error)
	}

	return resp.Data, nil
}

func (c *Client) ListSubscriptions() ([]pay.Subscription, error) {
	data, err := c.request("cent.subscription.list", nil)
	if err != nil {
		return nil, err
	}
	var subs []pay.Subscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

func (c *Client) ListSubscriptionsByUsername(username string) ([]pay.Subscription, error) {
	data, err := c.request("cent.subscription.list.username", []byte(username))
	if err != nil {
		return nil, err
	}
	var subs []pay.Subscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

func (c *Client) ListSubscriptionsByPlanID(planID int64) ([]pay.Subscription, error) {
	id := strconv.FormatInt(planID, 10)
	data, err := c.request("cent.subscription.list.plan_id", []byte(id))
	if err != nil {
		return nil, err
	}
	var subs []pay.Subscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

func (c *Client) GetSubscriptionsByCustomerID(customerID int64) (*pay.Subscription, error) {
	id := strconv.FormatInt(customerID, 10)
	data, err := c.request("cent.subscription.get.customer_id", []byte(id))
	if err != nil {
		return nil, err
	}
	var sub pay.Subscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (c *Client) GetSubscriptionsByProviderID(providerID string) (*pay.Subscription, error) {
	data, err := c.request("cent.subscription.get.provider_id", []byte(providerID))
	if err != nil {
		return nil, err
	}
	var sub pay.Subscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (c *Client) AddSubscriptionUser(su *pay.SubscriptionUser) error {
	data, err := json.Marshal(su)
	if err != nil {
		return err
	}

	_, err = c.request("cent.subscription.user.add", data)
	return err
}

func (c *Client) RemoveSubscriptionUser(su *pay.SubscriptionUser) error {
	data, err := json.Marshal(su)
	if err != nil {
		return err
	}

	_, err = c.request("cent.subscription.user.remove", data)
	return err
}

func (c *Client) ListSubscriptionUsers(subID int64) ([]string, error) {
	id := strconv.FormatInt(subID, 10)
	data, err := c.request("cent.subscription.user.list", []byte(id))
	if err != nil {
		return nil, err
	}
	var usernames []string
	if err := json.Unmarshal(data, &usernames); err != nil {
		return nil, err
	}
	return usernames, nil
}

func (c *Client) CountSubscriptionUsers(subID int64) (int64, error) {
	id := strconv.FormatInt(subID, 10)
	data, err := c.request("cent.subscription.user.count", []byte(id))
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(string(data), 10, 64)
}

func (c *Client) Sync() error {
	_, err := c.request("cent.sync", nil)
	return err

}

func (c *Client) Checkout(req *pay.CheckoutRequest) (string, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	data, err = c.request("cent.checkout", data)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type response struct {
	Success bool
	Error   string `json:",omitempty"`
	Data    []byte `json:",omitempty"`
}