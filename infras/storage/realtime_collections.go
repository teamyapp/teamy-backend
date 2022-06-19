package storage

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/infras/connection"
)

type Subscriptions map[uint64]*Subscription // <subscriptionID, subscription>
type Collection map[MutationType]Subscriptions

func newCollection() Collection {
	return map[MutationType]Subscriptions{
		CreateMutationType: map[uint64]*Subscription{},
		DeleteMutationType: map[uint64]*Subscription{},
		UpdateMutationType: map[uint64]*Subscription{},
	}
}

type Client struct {
	conn          connection.Connection
	subscriptions Subscriptions
}

type RealTimeCollections struct {
	nextClientID       uint64
	nextSubscriptionID uint64
	collections        map[string]Collection    // <collectionType, collection>
	subscriptions      map[uint64]*Subscription // <subscriptionID, subscription>
	clients            map[uint64]*Client       // <clientID, client>
}

func (r RealTimeCollections) RegisterCollectionType(collectionType string) {
	r.collections[collectionType] = newCollection()
}

func (r RealTimeCollections) ListCollectionTypes() []string {
	collectionTypes := make([]string, 0)
	for collectionType := range r.collections {
		collectionTypes = append(collectionTypes, collectionType)
	}

	return collectionTypes
}

func (r *RealTimeCollections) OnClientConnect(conn connection.Connection) uint64 {
	clientID := r.nextClientID
	r.clients[clientID] = &Client{
		conn:          conn,
		subscriptions: map[uint64]*Subscription{},
	}
	go func() {
		<-conn.OnClientDisconnect()
		r.onClientDisconnect(clientID)
	}()
	r.nextClientID++
	return clientID
}

func (r *RealTimeCollections) Subscribe(subscription Subscription) (uint64, error) {
	subscriptionID := r.nextSubscriptionID
	{
		client, ok := r.clients[subscription.ClientID]
		if !ok {
			return 0, fmt.Errorf("client not found: clientID=%v", subscription.ClientID)
		}
		client.subscriptions[subscriptionID] = &subscription
	}
	{
		collection, ok := r.collections[subscription.CollectionType]
		if !ok {
			return 0, fmt.Errorf("collection not found: collectionType=%v", subscription.CollectionType)
		}
		subscriptions := collection[subscription.MutationType]
		subscriptions[subscriptionID] = &subscription
	}
	{
		r.subscriptions[subscriptionID] = &subscription
	}
	r.nextSubscriptionID++
	return subscriptionID, nil
}

func (r RealTimeCollections) Unsubscribe(subscriptionID uint64) error {
	subscription, ok := r.subscriptions[subscriptionID]
	if !ok {
		return fmt.Errorf("subscription not found: subscriptionID=%v", subscriptionID)
	}

	{
		collection := r.collections[subscription.CollectionType]
		subscriptions := collection[subscription.MutationType]
		delete(subscriptions, subscriptionID)
	}
	{
		client := r.clients[subscription.ClientID]
		delete(client.subscriptions, subscriptionID)
	}
	{
		delete(r.subscriptions, subscriptionID)
	}

	return nil
}

func (r RealTimeCollections) ListSubscriptions() Subscriptions {
	return r.subscriptions
}

func (r RealTimeCollections) Mutate(mutation Mutation) error {
	log.Printf("new mutation: %v\n", mutation)
	collection, ok := r.collections[mutation.CollectionType]
	if !ok {
		return fmt.Errorf("collectionType not found: collectionType=%v", mutation.CollectionType)
	}

	subscriptions := collection[mutation.MutationType]
	for _, subscription := range subscriptions {
		match, err := subscription.match(mutation.Attributes)
		if err != nil {
			log.Println(err)
			return err
		}

		if !match {
			continue
		}

		err = r.notify(subscription.ClientID, mutation)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (r RealTimeCollections) onClientDisconnect(clientID uint64) {
	log.Printf("Client disconnect: clientID=%v\n", clientID)
	client := r.clients[clientID]
	for subscriptionID := range client.subscriptions {
		_ = r.Unsubscribe(subscriptionID)
	}
	delete(r.clients, clientID)
}

func (r RealTimeCollections) notify(clientID uint64, mutation Mutation) error {
	client := r.clients[clientID]
	buf, err := json.Marshal(mutation)
	if err != nil {
		return err
	}

	client.conn.SendMessage(buf)
	log.Printf("Notified mutation to client: clientID=%v, mutation%v\n", clientID, mutation)
	return nil
}

func NewRealTimeCollections() *RealTimeCollections {
	return &RealTimeCollections{
		nextClientID:       1,
		nextSubscriptionID: 1,
		collections:        map[string]Collection{},
		subscriptions:      map[uint64]*Subscription{},
		clients:            map[uint64]*Client{},
	}
}
