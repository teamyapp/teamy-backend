package realtime

import (
	"encoding/json"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const clientBufferSize = 50

type ClientNotifier struct {
	dataCollector               obs.DataCollector
	clientDisconnectSubscribers []chan bool
	messages                    chan entity.MessageEvent
}

func (c *ClientNotifier) subscribeClientDisconnect() <-chan bool {
	subscriber := make(chan bool)
	c.clientDisconnectSubscribers = append(c.clientDisconnectSubscribers, subscriber)
	return subscriber
}

func (c ClientNotifier) processMutation(message entity.MessageEvent) {
	c.messages <- message
}

func newClientNotifier(dataCollector obs.DataCollector, conn connection.Connection, clientID uint64) *ClientNotifier {
	messages := make(chan entity.MessageEvent, clientBufferSize)
	go func() {
		for message := range messages {
			var messagePayload entity.MessageEvent

			if message.Type == entity.MutationMessageType {
				mutation := message.Payload.(entity.MutationPayload)
				messagePayload = entity.MessageEvent{
					Type: message.Type,
					Payload: Mutation{
						CollectionType: mutation.CollectionType,
						MutationType:   mutation.MutationType,
						Payload:        mutation.Payload,
					},
				}
			} else {
				metadata := message.Payload.(entity.MetadataPayload)
				messagePayload = entity.MessageEvent{
					Type: message.Type,
					Payload: Metadata{
						ClientID: metadata.ClientID,
					},
				}
			}

			jsonBuf, err := json.MarshalIndent(messagePayload, "", "  ")
			if err != nil {
				dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
				continue
			}
			conn.SendMessage(jsonBuf)
			dataCollector.Logger.Log(obs.Info, obs.Props{
				obs.MessageProp: obs.Props{
					"summary":    "notification sent",
					"clientID":   clientID,
					"mutationID": message.ID,
				},
			})
		}
	}()
	clientNotifier := &ClientNotifier{
		clientDisconnectSubscribers: make([]chan bool, 0),
		messages:                    messages,
	}
	go func() {
		<-conn.OnClientDisconnect()
		for _, subscriber := range clientNotifier.clientDisconnectSubscribers {
			subscriber <- true
		}
	}()
	return clientNotifier
}
