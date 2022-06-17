package stream

type Subscription[Item any] struct {
	pubSub PubSub[Item]
	output chan Item
}

func (s *Subscription[Item]) Unsubscribe() {
	s.pubSub.unsubscribe(s)
	close(s.output)
}

func (s *Subscription[Item]) Output() <-chan Item {
	return s.output
}

func newSubscription[Item any](pubSub PubSub[Item]) *Subscription[Item] {
	return &Subscription[Item]{
		pubSub: pubSub,
		output: make(chan Item),
	}
}

type PubSub[Item any] struct {
	subscriptions map[*Subscription[Item]]bool
}

func (p PubSub[Item]) unsubscribe(subscription *Subscription[Item]) {
	delete(p.subscriptions, subscription)
}

func (p PubSub[Item]) Subscribe() *Subscription[Item] {
	subscription := newSubscription[Item](p)
	p.subscriptions[subscription] = true
	return subscription
}

func NewPubSub[Item any](input <-chan Item) *PubSub[Item] {
	pubSub := &PubSub[Item]{
		subscriptions: make(map[*Subscription[Item]]bool),
	}
	go func() {
		for item := range input {
			for subscription := range pubSub.subscriptions {
				go func(sub *Subscription[Item]) {
					sub.output <- item
				}(subscription)
			}
		}
		for subscription := range pubSub.subscriptions {
			subscription.Unsubscribe()
		}
	}()
	return pubSub
}
