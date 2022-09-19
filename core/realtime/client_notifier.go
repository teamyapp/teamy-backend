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
        clientID uint64
	messages                    chan MessageEvent
	acceptMutation                     bool
}

func (c *ClientNotifier) subscribeClientDisconnect() <-chan bool {
	subscriber := make(chan bool)
	c.clientDisconnectSubscribers = append(c.clientDisconnectSubscribers, subscriber)
	return subscriber
}

func (c *ClientNotifier) onInitialStateReady(isReady bool) {
	c.acceptMutation = true
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

func (c *ClientNotifier) sentMetadata(clientID uint64) {
	message := MessageEvent{
		Type: MetadataMessageType,
		Payload: MetadataMessage{
			ClientID: c.clientID,
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
                 clientID: clientID,
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
