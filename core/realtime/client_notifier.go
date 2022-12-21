package realtime

import (
	"context"
	"encoding/json"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
)

const clientBufferSize = 50

type ClientNotifier struct {
	dataCollector               obs.DataCollector
	clientDisconnectSubscribers []chan bool
	clientID                    uint64
	messages                    chan Message
	acceptTransaction           bool
}

func (c *ClientNotifier) subscribeClientDisconnect() <-chan bool {
	subscriber := make(chan bool)
	c.clientDisconnectSubscribers = append(c.clientDisconnectSubscribers, subscriber)
	return subscriber
}

func (c *ClientNotifier) onInitialStateReady() {
	c.acceptTransaction = true
}

func (c *ClientNotifier) notifyTransaction(ct context.Context, transaction *Transaction) {
	c.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"Summary": "process transaction",
		},
	})

	if !c.acceptTransaction {
		c.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
			obs.MessageProp: obs.Props{
				"Summary": "discard transaction",
			},
		})
		return
	}

	mutationMessages := collect.Map(transaction.mutations, func(mutation Mutation, _ int) MutationMessage {
		return MutationMessage{
			ID:             mutation.ID,
			CollectionType: mutation.CollectionType,
			MutationType:   mutation.MutationType,
			Payload:        mutation.Payload,
		}
	})

	transactionMessage := TransactionMessage{
		ID:        transaction.id,
		TeamID:    transaction.teamID,
		Mutations: mutationMessages,
	}

	message := Message{
		Type:    TransactionMessageType,
		Payload: transactionMessage,
	}
	c.messages <- message
}

func (c *ClientNotifier) sentMetadata() {
	message := Message{
		Type: MetadataMessageType,
		Payload: MetadataMessage{
			ClientID: c.clientID,
		},
	}
	c.messages <- message
}

func newClientNotifier(dataCollector obs.DataCollector, conn connection.Connection, clientID uint64) *ClientNotifier {
	messages := make(chan Message, clientBufferSize)
	ct := context.Background()
	ct = ctx.WithClientID(ct, clientID)
	go func() {
		for message := range messages {
			jsonBuf, err := json.MarshalIndent(message, "", "  ")

			if err != nil {
				dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				continue
			}
			conn.SendMessage(jsonBuf)
			dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
				obs.MessageProp: obs.Props{
					"Summary": "notification sent",
				},
			})
		}
	}()
	clientNotifier := &ClientNotifier{
		dataCollector:               dataCollector,
		clientDisconnectSubscribers: make([]chan bool, 0),
		clientID:                    clientID,
		messages:                    messages,
		acceptTransaction:           false,
	}
	go func() {
		<-conn.OnClientDisconnect()
		for _, subscriber := range clientNotifier.clientDisconnectSubscribers {
			subscriber <- true
		}
	}()
	return clientNotifier
}
