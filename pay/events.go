package pay

type events struct {
	subAddedCallbacks        []func(*Subscription)
	subUpdatedCallbacks      []func(*Subscription, *Subscription)
	subRemovedCallbacks      []func(*Subscription)
	seatAddedCallbacks       []func(*Subscription, string)
	seatRemovedCallbacks     []func(*Subscription, string)
	customerAddedCallbacks   []func(*Customer)
	customerUpdatedCallbacks []func(*Customer, *Customer)
	customerRemovedCallbacks []func(*Customer)
	planAddedCallbacks       []func(*Plan)
	planUpdatedCallbacks     []func(*Plan, *Plan)
	planRemovedCallbacks     []func(*Plan)
	priceAddedCallbacks      []func(*Price)
	priceUpdatedCallbacks    []func(*Price, *Price)
	priceRemovedCallbacks    []func(*Price)
}

func (e *events) OnSeatAdded(cb func(*Subscription, string)) {
	e.seatAddedCallbacks = append(e.seatAddedCallbacks, cb)
}

func (e *events) OnSeatRemoved(cb func(*Subscription, string)) {
	e.seatRemovedCallbacks = append(e.seatRemovedCallbacks, cb)
}

func (e *events) OnSubscriptionAdded(cb func(*Subscription)) {
	e.subAddedCallbacks = append(e.subAddedCallbacks, cb)
}

func (e *events) OnSubscriptionUpdated(cb func(*Subscription, *Subscription)) {
	e.subUpdatedCallbacks = append(e.subUpdatedCallbacks, cb)
}

func (e *events) OnSubscriptionRemoved(cb func(*Subscription)) {
	e.subRemovedCallbacks = append(e.subRemovedCallbacks, cb)
}

func (e *events) OnCustomerAdded(cb func(*Customer)) {
	e.customerAddedCallbacks = append(e.customerAddedCallbacks, cb)
}

func (e *events) OnCustomerUpdated(cb func(*Customer, *Customer)) {
	e.customerUpdatedCallbacks = append(e.customerUpdatedCallbacks, cb)
}

func (e *events) OnCustomerRemoved(cb func(*Customer)) {
	e.customerRemovedCallbacks = append(e.customerRemovedCallbacks, cb)
}

func (e *events) OnPlanAdded(cb func(*Plan)) {
	e.planAddedCallbacks = append(e.planAddedCallbacks, cb)
}

func (e *events) OnPlanUpdated(cb func(*Plan, *Plan)) {
	e.planUpdatedCallbacks = append(e.planUpdatedCallbacks, cb)
}

func (e *events) OnPlanRemoved(cb func(*Plan)) {
	e.planRemovedCallbacks = append(e.planRemovedCallbacks, cb)
}

func (e *events) OnPriceAdded(cb func(*Price)) {
	e.priceAddedCallbacks = append(e.priceAddedCallbacks, cb)
}

func (e *events) OnPriceUpdated(cb func(*Price, *Price)) {
	e.priceUpdatedCallbacks = append(e.priceUpdatedCallbacks, cb)
}

func (e *events) OnPriceRemoved(cb func(*Price)) {
	e.priceRemovedCallbacks = append(e.priceRemovedCallbacks, cb)
}

func (e *events) subAdded(s *Subscription) {
	for _, cb := range e.subAddedCallbacks {
		cb(s)
	}
}

func (e *events) subUpdated(prev *Subscription, s *Subscription) {
	for _, cb := range e.subUpdatedCallbacks {
		cb(prev, s)
	}
}

func (e *events) subRemoved(s *Subscription) {
	for _, cb := range e.subRemovedCallbacks {
		cb(s)
	}
}

func (e *events) customerAdded(c *Customer) {
	for _, cb := range e.customerAddedCallbacks {
		cb(c)
	}
}

func (e *events) customerUpdated(prev *Customer, s *Customer) {
	for _, cb := range e.customerUpdatedCallbacks {
		cb(prev, s)
	}
}

func (e *events) customerRemoved(c *Customer) {
	for _, cb := range e.customerRemovedCallbacks {
		cb(c)
	}
}

func (e *events) planAdded(c *Plan) {
	for _, cb := range e.planAddedCallbacks {
		cb(c)
	}
}

func (e *events) planUpdated(prev *Plan, s *Plan) {
	for _, cb := range e.planUpdatedCallbacks {
		cb(prev, s)
	}
}

func (e *events) planRemoved(c *Plan) {
	for _, cb := range e.planRemovedCallbacks {
		cb(c)
	}
}

func (e *events) priceAdded(c *Price) {
	for _, cb := range e.priceAddedCallbacks {
		cb(c)
	}
}

func (e *events) priceUpdated(prev *Price, s *Price) {
	for _, cb := range e.priceUpdatedCallbacks {
		cb(prev, s)
	}
}

func (e *events) priceRemoved(c *Price) {
	for _, cb := range e.priceRemovedCallbacks {
		cb(c)
	}
}

func (e *events) seatAdded(s *Subscription, seat string) {
	for _, cb := range e.seatAddedCallbacks {
		cb(s, seat)
	}
}

func (e *events) seatRemoved(s *Subscription, seat string) {
	for _, cb := range e.seatRemovedCallbacks {
		cb(s, seat)
	}
}
