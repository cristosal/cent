package cent

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cristosal/cent/pay"
	"github.com/nats-io/nats.go"
)

var ErrBadRequest = errors.New("bad request")

type Server struct {
	nc       *nats.Conn
	js       nats.JetStreamContext
	provider *pay.StripeProvider
	cfg      *Config
}

type Config struct {
	NatsURL         string
	Queue           string
	WebhookEndpoint string
	EnableWebUI     bool
	Provider        *pay.StripeProvider
	HttpAddr        string
}

func (cfg *Config) setDefaults() {
	if cfg.HttpAddr == "" {
		cfg.HttpAddr = "127.0.0.1:8080"
	}

	if cfg.Queue == "" {
		cfg.Queue = "cent"
	}

	if cfg.NatsURL == "" {
		cfg.NatsURL = nats.DefaultURL
	}

	if cfg.WebhookEndpoint == "" {
		cfg.WebhookEndpoint = "/webhook"
	}
}

func (s *Server) Listen() error {
	if s.cfg.Provider == nil {
		return fmt.Errorf("provider is required")
	}

	nc, err := nats.Connect(s.cfg.NatsURL)
	if err != nil {
		return fmt.Errorf("error connecting to nats: %w", err)
	}
	s.nc = nc

	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("error initializing jet stream: %w", err)
	}

	s.js = js

	s.forwardProviderEvents()

	if err := s.registerNATSHandlers(); err != nil {
		return err
	}

	s.registerHTTPHandlers()
	return http.ListenAndServe(s.cfg.HttpAddr, nil)
}

func New(cfg *Config) *Server {
	cfg.setDefaults()
	srv := Server{
		provider: cfg.Provider,
		cfg:      cfg,
	}

	return &srv
}

func (s *Server) registerHTTPHandlers() {
	http.HandleFunc(s.cfg.WebhookEndpoint, s.cfg.Provider.Webhook())
	if s.cfg.EnableWebUI {
		handleWebUI(s.cfg.Provider, s.cfg.HttpAddr)
	}
}

func (s *Server) registerNATSHandlers() error {
	submap := map[string]natsHandler{
		SubjCheckout:                     s.handleCheckout(),
		SubjCustomerAdd:                  s.handleAddCustomer(),
		SubjCustomerGetByEmail:           s.handleGetCustomerByEmail(),
		SubjCustomerGetByID:              s.handleGetCustomerByID(),
		SubjCustomerGetByProviderID:      s.handleGetCustomerByProvider(),
		SubjCustomerList:                 s.handleListCustomers(),
		SubjCustomerRemoveByProviderID:   s.handleRemoveCustomerByProviderID(),
		SubjCustomerUpdate:               s.handleUpdateCustomer(),
		SubjPlanAdd:                      s.handleAddPlan(),
		SubjPlanGetByID:                  s.handleGetPlanByID(),
		SubjPlanGetByName:                s.handleGetPlanByName(),
		SubjPlanGetByPriceID:             s.handleGetPlanByPriceID(),
		SubjPlanGetByProviderID:          s.handleGetPlanByProviderID(),
		SubjPlanGetBySubscriptionID:      s.handleGetPlanBySubscriptionID(),
		SubjPlanList:                     s.handleListPlans(),
		SubjPlanListActive:               s.handleListActivePlans(),
		SubjPlanListByUsername:           s.handleGetPlansByUsername(),
		SubjPlanRemoveByProviderID:       s.handleRemovePlanByProviderID(),
		SubjPlanUpdate:                   s.handleUpdatePlan(),
		SubjPriceAdd:                     s.handleAddPrice(),
		SubjPriceGetByID:                 s.handleGetPriceByID(),
		SubjPriceGetByProviderID:         s.handleGetPriceByProviderID(),
		SubjPriceList:                    s.handleListPrices(),
		SubjPriceListByPlanID:            s.handleListPricesByPlanID(),
		SubjSubscriptionGetByID:          s.handleGetSubscriptionByID(),
		SubjSubscriptionGetByProviderID:  s.handleGetSubscriptionByProviderID(),
		SubjSubscriptionList:             s.handleListSubscriptions(),
		SubjSubscriptionListByCustomerID: s.handleListSubscriptionsByCustomerID(),
		SubjSubscriptionListByPlanID:     s.handleListSubscriptionsByPlanID(),
		SubjSubscriptionListByUsername:   s.handleListSubscriptionsByUsername(),
		SubjSubscriptionUserAdd:          s.handleAddSubscriptionUser(),
		SubjSubscriptionUserCount:        s.handleCountSubscriptionUsers(),
		SubjSubscriptionUserList:         s.handleListSubscriptionUsers(),
		SubjSubscriptionUserRemove:       s.handleRemoveSubscriptionUser(),
		SubjSync:                         s.handleSync(),
	}

	for k, v := range submap {
		_, err := s.sub(k, v)
		if err != nil {
			return fmt.Errorf("error subscribing to %s: %w", k, err)
		}
	}

	return nil
}

func (ns *Server) forwardProviderEvents() {
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
	p.OnCustomerAdded(func(c *pay.Customer) {
		pub(SubjCustomerAdded, c)
	})

	p.OnCustomerRemoved(func(c *pay.Customer) {
		pub(SubjCustomerRemoved, c)
	})

	p.OnCustomerUpdated(func(_, c2 *pay.Customer) {
		pub(SubjCustomerUpdated, c2)
	})

	p.OnSubscriptionAdded(func(s *pay.Subscription) {
		pub(SubjSubscriptionAdded, s)
		pub(SubjSubscriptionActivated, s)
	})

	p.OnSubscriptionRemoved(func(s *pay.Subscription) {
		pub(SubjSubscriptionRemoved, s)
		pub(SubjSubscriptionDeactivated, s)
	})

	p.OnSubscriptionUpdated(func(previous, current *pay.Subscription) {
		pub(SubjSubscriptionUpdated, current)
		if previous.Active && !current.Active {
			pub(SubjSubscriptionDeactivated, current)
		} else if !previous.Active && current.Active {
			pub(SubjSubscriptionActivated, current)
		}
	})

	p.OnPlanAdded(func(p *pay.Plan) {
		pub(SubjPlanAdded, p)
	})

	p.OnPlanRemoved(func(p *pay.Plan) {
		pub(SubjPlanRemoved, p)
	})

	p.OnPlanUpdated(func(_ *pay.Plan, p2 *pay.Plan) {
		pub(SubjPlanUpdated, p2)
	})

	p.OnPriceAdded(func(p *pay.Price) {
		pub(SubjPriceAdded, p)
	})

	p.OnPriceRemoved(func(p *pay.Price) {
		pub(SubjPriceRemoved, p)
	})

	p.OnPriceUpdated(func(_ *pay.Price, p2 *pay.Price) {
		pub(SubjPriceUpdated, p2)
	})

	p.OnSeatAdded(func(s1 *pay.Subscription, username string) {
		pub(SubjSubscriptionUserAdded, pay.SubscriptionUser{
			SubscriptionID: s1.ID,
			Username:       username,
		})
	})

	p.OnSeatRemoved(func(s1 *pay.Subscription, username string) {
		pub(SubjSubscriptionUserRemoved, pay.SubscriptionUser{
			SubscriptionID: s1.ID,
			Username:       username,
		})
	})
}

// ---------------------------------------------------
func (s *Server) handleAddCustomer() natsHandler {
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

func (s *Server) handleGetCustomerByEmail() natsHandler {
	return func(msg *nats.Msg) error {
		c, err := s.provider.GetCustomerByEmail(string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, c)
	}
}

func (s *Server) handleGetCustomerByID() natsHandler {
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

func (s *Server) handleGetCustomerByProvider() natsHandler {
	return func(msg *nats.Msg) error {
		c, err := s.provider.GetCustomerByProvider(pay.ProviderStripe, string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, c)
	}
}

func (s *Server) handleListCustomers() natsHandler {
	return func(msg *nats.Msg) error {
		customers, err := s.provider.ListAllCustomers()
		if err != nil {
			return err
		}

		return s.reply(msg, customers)
	}
}

func (s *Server) handleRemoveCustomerByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		if err := s.provider.RemoveCustomerByProviderID(string(msg.Data)); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *Server) handleUpdateCustomer() natsHandler {
	return func(msg *nats.Msg) error {
		var cust pay.Customer
		if err := json.Unmarshal(msg.Data, &cust); err != nil {
			return ErrBadRequest
		}

		if err := s.provider.UpdateCustomer(&cust); err != nil {
			return err
		}

		return s.reply(msg, &cust)
	}
}

// ---------------------------------------------------
func (s *Server) handleAddPlan() natsHandler {
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

func (s *Server) handleGetPlanByID() natsHandler {
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

func (s *Server) handleGetPlanByPriceID() natsHandler {
	return func(msg *nats.Msg) error {
		id, err := strconv.ParseInt(string(msg.Data), 10, 64)
		if err != nil {
			return ErrBadRequest
		}

		pl, err := s.provider.GetPlanByPriceID(id)
		if err != nil {
			return err
		}

		return s.reply(msg, pl)
	}
}

func (s *Server) handleGetPlanBySubscriptionID() natsHandler {
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

func (s *Server) handleGetPlanByName() natsHandler {
	return func(msg *nats.Msg) error {
		pl, err := s.provider.GetPlanByName(string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, pl)
	}
}

func (s *Server) handleGetPlanByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		pl, err := s.provider.GetPlanByProviderID(pay.ProviderStripe, string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, pl)
	}
}

func (s *Server) handleGetPlansByUsername() natsHandler {
	return func(msg *nats.Msg) error {
		plans, err := s.provider.GetPlansByUsername(string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, plans)
	}
}

func (s *Server) handleListPlans() natsHandler {
	return func(msg *nats.Msg) error {
		plans, err := s.provider.ListPlans()
		if err != nil {
			return err
		}

		return s.reply(msg, plans)
	}
}

func (s *Server) handleListActivePlans() natsHandler {
	return func(msg *nats.Msg) error {
		plans, err := s.provider.ListActivePlans()
		if err != nil {
			return err
		}

		return s.reply(msg, plans)
	}
}

func (s *Server) handleRemovePlanByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		if err := s.provider.RemovePlanByProviderID(string(msg.Data)); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

func (s *Server) handleUpdatePlan() natsHandler {
	return func(msg *nats.Msg) error {
		var p pay.Plan
		if err := json.Unmarshal(msg.Data, &p); err != nil {
			return ErrBadRequest
		}

		if err := s.provider.UpdatePlan(&p); err != nil {
			return err
		}

		return s.reply(msg, &p)
	}
}

// ---------------------------------------------------------
func (s *Server) handleListSubscriptions() natsHandler {
	return func(msg *nats.Msg) error {
		subs, err := s.provider.ListAllSubscriptions()
		if err != nil {
			return err
		}

		return s.reply(msg, subs)
	}
}

func (s *Server) handleListSubscriptionsByUsername() natsHandler {
	return func(msg *nats.Msg) error {
		subs, err := s.provider.ListSubscriptionsByUsername(string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, subs)
	}
}

func (s *Server) handleListSubscriptionsByPlanID() natsHandler {
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

func (s *Server) handleListSubscriptionsByCustomerID() natsHandler {
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

func (s *Server) handleGetSubscriptionByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		sub, err := s.provider.GetSubscriptionByProvider(pay.ProviderStripe, string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, sub)
	}
}

func (s *Server) handleGetSubscriptionByID() natsHandler {
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

// ------------------------------------------------------------
func (s *Server) handleListPrices() natsHandler {
	return func(msg *nats.Msg) error {
		prices, err := s.provider.ListAllPrices()
		if err != nil {
			return err
		}

		return s.reply(msg, prices)
	}
}

func (s *Server) handleListPricesByPlanID() natsHandler {
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

func (s *Server) handleGetPriceByID() natsHandler {
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

func (s *Server) handleGetPriceByProviderID() natsHandler {
	return func(msg *nats.Msg) error {
		pr, err := s.provider.GetPriceByProvider(pay.ProviderStripe, string(msg.Data))
		if err != nil {
			return err
		}

		return s.reply(msg, pr)
	}
}

func (s *Server) handleAddPrice() natsHandler {
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

// -------------------------------------------------------------
func (s *Server) handleAddSubscriptionUser() natsHandler {
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

func (s *Server) handleCountSubscriptionUsers() natsHandler {
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

func (s *Server) handleListSubscriptionUsers() natsHandler {
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

func (s *Server) handleRemoveSubscriptionUser() natsHandler {
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

// ------------------------------------------------------------

func (s *Server) handleCheckout() natsHandler {
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

func (s *Server) handleSync() natsHandler {
	return func(msg *nats.Msg) error {
		if err := s.provider.Sync(); err != nil {
			return err
		}

		return s.reply(msg, nil)
	}
}

// ------------------------------------------------------------
type natsHandler func(msg *nats.Msg) error

func (Server) reply(msg *nats.Msg, v any) error {
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

func (s *Server) sub(subj string, h natsHandler) (*nats.Subscription, error) {
	return s.nc.QueueSubscribe(subj, s.cfg.Queue, func(msg *nats.Msg) {
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
