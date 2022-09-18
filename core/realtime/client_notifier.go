package realtime

import (
	"encoding/json"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/obs"
)

const clientBufferSize = 50

type ClientNotifier struct {
	dataCollector               obs.DataCollector
	clientDisconnectSubscribers []chan bool
	messages                    chan MessageEvent
	isReady                     bool
}

func (c *ClientNotifier) subscribeClientDisconnect() <-chan bool {
	subscriber := make(chan bool)
	c.clientDisconnectSubscribers = append(c.clientDisconnectSubscribers, subscriber)
	return subscriber
}

func (c *ClientNotifier) setClientIsReady(isReady bool) {
	c.isReady = isReady
}

func (c *ClientNotifier) processMutation(mutation Mutation) {
	if !c.isReady {
		return
	}

	message := MessageEvent{
		Type: MutationMessageType,
		Payload: MutationMessage{
			CollectionType: mutation.CollectionType,
			MutationType:   mutation.MutationType,
			Payload:        mutation.Payload,
		},
	}

	c.messages <- message
}

func (c *ClientNotifier) sendClientID(clientID uint64) {
	message := MessageEvent{
		Type: MetadataMessageType,
		Payload: MetadataMessage{
			ClientID: clientID,
		},
	}

	c.messages <- message
}

func newClientNotifier(dataCollector obs.DataCollector, conn connection.Connection, clientID uint64) *ClientNotifier {
	messages := make(chan MessageEvent, clientBufferSize)
	go func() {
		for message := range messages {
			jsonBuf, err := json.MarshalIndent(message, "", "  ")
			if err != nil {
				dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
				continue
			}
			conn.SendMessage(jsonBuf)
			dataCollector.Logger.Log(obs.Info, obs.Props{
				obs.MessageProp: obs.Props{
					"summary":  "notification sent",
					"clientID": clientID,
				},
			})
		}
	}()
	clientNotifier := &ClientNotifier{
		clientDisconnectSubscribers: make([]chan bool, 0),
		messages:                    messages,
		isReady:                     false,
	}
	go func() {
		<-conn.OnClientDisconnect()
		for _, subscriber := range clientNotifier.clientDisconnectSubscribers {
			subscriber <- true
		}
	}()
	return clientNotifier
}
