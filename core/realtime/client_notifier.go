package realtime

import (
	"context"
	"encoding/json"

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
	acceptMutation              bool
}

func (c *ClientNotifier) subscribeClientDisconnect() <-chan bool {
	subscriber := make(chan bool)
	c.clientDisconnectSubscribers = append(c.clientDisconnectSubscribers, subscriber)
	return subscriber
}

func (c *ClientNotifier) onInitialStateReady() {
	c.acceptMutation = true
}

func (c *ClientNotifier) processMutation(mutation Mutation) {
	ct := context.Background()
	ct = ctx.WithClientID(ct, c.clientID)
	ct = WithMutationID(ct, mutation.ID)
	c.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"summary": "process mutation",
		},
	})

	if !c.acceptMutation {
		c.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
			obs.MessageProp: obs.Props{
				"summary": "discard mutation",
			},
		})
		return
	}

	message := Message{
		Type: MutationMessageType,
		Payload: MutationMessage{
			CollectionType: mutation.CollectionType,
			MutationType:   mutation.MutationType,
			Payload:        mutation.Payload,
		},
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
					"summary": "notification sent",
				},
			})
		}
	}()
	clientNotifier := &ClientNotifier{
		dataCollector:               dataCollector,
		clientDisconnectSubscribers: make([]chan bool, 0),
		clientID:                    clientID,
		messages:                    messages,
		acceptMutation:              false,
	}
	go func() {
		<-conn.OnClientDisconnect()
		for _, subscriber := range clientNotifier.clientDisconnectSubscribers {
			subscriber <- true
		}
	}()
	return clientNotifier
}
