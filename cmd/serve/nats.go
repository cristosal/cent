package serve

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/cristosal/cent/client"
	"github.com/cristosal/pay"
	"github.com/nats-io/nats.go"
)

var ErrBadRequest = errors.New("bad request")

type natsServer struct {
	nc       *nats.Conn
	js       nats.JetStreamContext
	provider *pay.StripeProvider
	queue    string
}

type natsServerConfig struct {
	NatsURL  string
	Queue    string
	Provider *pay.StripeProvider
}

func newNatsServer(cfg *natsServerConfig) (*natsServer, error) {
	if cfg.Provider == nil {
		return nil, fmt.Errorf("provider is required")
	}

	if cfg.Queue == "" {
		cfg.Queue = "cent"
	}

	if cfg.NatsURL == "" {
		cfg.NatsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("error initializing jet stream: %w", err)
	}

	p := cfg.Provider

	srv := natsServer{
		nc:       nc,
		js:       js,
		provider: p,
	}

	srv.attachProviderEvents()
	return &srv, nil
}

func (ns *natsServer) attachProviderEvents() {
	pub := func(subj string, v any) error {
		data, err := json.Marshal(v)
		if err != nil {
			fmt.Printf("error marshaling data for publish on subj %s: %v\n", subj, err)
			return err
		}

		_, err = ns.js.Publish(subj, data)
		return err
	}

	p := ns.provider
	p.OnCustomerAdded(func(c *pay.Customer) { pub(client.SubjCustomerAdded, c) })
	p.OnCustomerRemoved(func(c *pay.Customer) { pub(client.SubjCustomerRemoved, c) })
	p.OnCustomerUpdated(func(_, c2 *pay.Customer) { pub(client.SubjCustomerUpdated, c2) })

	p.OnSubscriptionAdded(func(s *pay.Subscription) {
		pub(client.SubjSubscriptionAdded, s)
		pub(client.SubjSubscriptionActivated, s)
	})

	p.OnSubscriptionRemoved(func(s *pay.Subscription) {
		pub(client.SubjSubscriptionRemoved, s)
		pub(client.SubjSubscriptionDeactivated, s)
	})

	p.OnSubscriptionUpdated(func(s1, s2 *pay.Subscription) {
		pub(client.SubjSubscriptionUpdated, s2)
		if s1.Active && !s2.Active {
			pub(client.SubjSubscriptionDeactivated, s2)
		} else if !s1.Active && s2.Active {
			pub(client.SubjSubscriptionActivated, s2)
		}
	})

	p.OnPlanAdded(func(p *pay.Plan) { pub(client.SubjPlanAdded, p) })
	p.OnPlanRemoved(func(p *pay.Plan) { pub(client.SubjPlanRemoved, p) })
	p.OnPlanUpdated(func(_ *pay.Plan, p2 *pay.Plan) { pub(client.SubjPlanUpdated, p2) })
	p.OnPriceAdded(func(p *pay.Price) { pub(client.SubjPriceAdded, p) })
	p.OnPriceRemoved(func(p *pay.Price) { pub(client.SubjPriceRemoved, p) })
	p.OnPriceUpdated(func(_ *pay.Price, p2 *pay.Price) { pub(client.SubjPriceUpdated, p2) })
}

func (s *natsServer) handleAddCustomer() natsHandler {
	return func(msg *nats.Msg) error {
		var c pay.Customer
		if err := json.Unmarshal(msg.Data, &c); err != nil {
			return err
		}

		if err := s.provider.AddCustomer(&c); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *natsServer) handleRemoveCustomerByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		if err := s.provider.RemoveCustomerByProviderID(string(msg.Data)); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *natsServer) handleListCustomers() natsHandler {
	return func(msg *nats.Msg) error {
		customers, err := s.provider.ListAllCustomers()
		if err != nil {
			return err
		}

		return s.reply(msg, customers)
	}
}

func (s *natsServer) handleGetCustomerByID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}

		c, err := s.provider.GetCustomerByID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, c)
	}
}

func (s *natsServer) handleGetCustomerByEmail() natsHandler {
	return func(msg *nats.Msg) error {
		c, err := s.provider.GetCustomerByEmail(string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, c)
	}
}

func (s *natsServer) handleGetCustomerByProvider() natsHandler {
	return func(msg *nats.Msg) error {
		c, err := s.provider.GetCustomerByProvider(pay.ProviderStripe, string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, c)
	}
}

func (s *natsServer) handleAddPlan() natsHandler {
	return func(msg *nats.Msg) error {
		var pl pay.Plan
		if err := json.Unmarshal(msg.Data, &pl); err != nil {
			return err
		}
		if err := s.provider.AddPlan(&pl); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *natsServer) handleGetPlanByID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}

		pl, err := s.provider.GetPlanByID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, pl)
	}
}

func (s *natsServer) handleGetPlanBySubscriptionID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}

		pl, err := s.provider.GetPlanBySubscriptionID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, pl)
	}
}

func (s *natsServer) handleGetPlanByName() natsHandler {
	return func(msg *nats.Msg) error {
		pl, err := s.provider.GetPlanByName(string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, pl)
	}
}

func (s *natsServer) handleGetPlanByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		pl, err := s.provider.GetPlanByProviderID(pay.ProviderStripe, string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, pl)
	}
}

func (s *natsServer) handleGetPlansByUsername() natsHandler {
	return func(msg *nats.Msg) error {
		plans, err := s.provider.GetPlansByUsername(string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, plans)
	}
}

func (s *natsServer) handleListPlans() natsHandler {
	return func(msg *nats.Msg) error {
		plans, err := s.provider.ListPlans()
		if err != nil {
			return err
		}

		return s.reply(msg, plans)
	}
}

func (s *natsServer) handleListActivePlans() natsHandler {
	return func(msg *nats.Msg) error {
		plans, err := s.provider.ListActivePlans()
		if err != nil {
			return err
		}

		return s.reply(msg, plans)
	}
}

func (s *natsServer) handleRemovePlanByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		if err := s.provider.RemovePlanByProviderID(string(msg.Data)); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *natsServer) handleListSubscriptions() natsHandler {
	return func(msg *nats.Msg) error {
		subs, err := s.provider.ListAllSubscriptions()
		if err != nil {
			return err
		}

		return s.reply(msg, subs)
	}
}

func (s *natsServer) handleListSubscriptionsByUsername() natsHandler {
	return func(msg *nats.Msg) error {
		subs, err := s.provider.ListSubscriptionsByUsername(string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, subs)
	}
}

func (s *natsServer) handleListSubscriptionsByPlanID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}

		subs, err := s.provider.ListSubscriptionsByPlanID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, subs)
	}
}

func (s *natsServer) handleListSubscriptionsByCustomerID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}

		subs, err := s.provider.ListSubscriptionsByCustomerID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, subs)
	}
}

func (s *natsServer) handleGetSubscriptionByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		sub, err := s.provider.GetSubscriptionByProvider(pay.ProviderStripe, string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, sub)
	}
}

func (s *natsServer) handleGetSubscriptionByID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return err
		}

		sub, err := s.provider.GetSubscriptionByID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, sub)
	}
}

func (s *natsServer) handleListPrices() natsHandler {
	return func(msg *nats.Msg) error {
		prices, err := s.provider.ListAllPrices()
		if err != nil {
			return err
		}

		return s.reply(msg, prices)
	}
}

func (s *natsServer) handleListPricesByPlanID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}

		subs, err := s.provider.ListPricesByPlanID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, subs)
	}
}

func (s *natsServer) handleGetPriceByID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}

		pr, err := s.provider.GetPriceByID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, pr)
	}
}

func (s *natsServer) handleGetPriceByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		pr, err := s.provider.GetPriceByProvider(pay.ProviderStripe, string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, pr)
	}
}

func (s *natsServer) handleAddPrice() natsHandler {
	return func(msg *nats.Msg) error {
		var pr pay.Price
		if err := json.Unmarshal(msg.Data, &pr); err != nil {
			return err
		}
		if err := s.provider.AddPrice(&pr); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *natsServer) handleAddSubscriptionUser() natsHandler {
	return func(msg *nats.Msg) error {
		var su pay.SubscriptionUser
		if err := json.Unmarshal(msg.Data, &su); err != nil {
			return err
		}
		if err := s.provider.AddSubscriptionUser(&su); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *natsServer) handleRemoveSubscriptionUser() natsHandler {
	return func(msg *nats.Msg) error {
		var su pay.SubscriptionUser
		if err := json.Unmarshal(msg.Data, &su); err != nil {
			return err
		}
		if err := s.provider.RemoveSubscriptionUser(&su); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *natsServer) handleListSubscriptionUsers() natsHandler {
	return func(msg *nats.Msg) error {
		subID, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}
		usernames, err := s.provider.ListUsernames(subID)
		if err != nil {
			return err
		}

		return s.reply(msg, usernames)
	}
}

func (s *natsServer) handleCountSubscriptionUsers() natsHandler {
	return func(msg *nats.Msg) error {
		subID, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}
		count, err := s.provider.CountSubscriptionUsers(subID)
		if err != nil {
			return err
		}

		return s.reply(msg, count)
	}
}

func (s *natsServer) handleCheckout() natsHandler {
	return func(msg *nats.Msg) error {
		var req pay.CheckoutRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return err
		}

		url, err := s.provider.Checkout(&req)
		if err != nil {
			return nil
		}

		return s.reply(msg, url)
	}
}

func (s *natsServer) handleSync() natsHandler {
	return func(msg *nats.Msg) error {
		if err := s.provider.Sync(); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

type natsHandler func(msg *nats.Msg) error

func (natsServer) reply(msg *nats.Msg, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	res := response{
		Success: true,
		Data:    data,
	}

	resdata, err := json.Marshal(&res)
	if err != nil {
		return err
	}

	return msg.Respond(resdata)
}

func (s *natsServer) sub(subj string, h natsHandler) (*nats.Subscription, error) {
	return s.nc.QueueSubscribe(subj, s.queue, func(msg *nats.Msg) {
		if err := h(msg); err != nil {
			data, err := json.Marshal(&response{
				Success: false,
				Error:   err.Error(),
			})
			if err != nil {
				// this is fatal we should never get here
				err = fmt.Errorf("error marshaling resonse json: %w", err)
				panic(err)
			}

			msg.Respond(data)
		}
	})
}

type response struct {
	Success bool
	Error   string `json:",omitempty"`
	Data    []byte `json:",omitempty"`
}
