package realtime

import (
	"encoding/json"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/obs"
)

const clientBufferSize = 50

type MutationMessage struct {
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
}

type ClientNotifier struct {
	dataCollector               obs.DataCollector
	clientDisconnectSubscribers []chan bool
	mutations                   chan Mutation
}

func (c *ClientNotifier) subscribeClientDisconnect() <-chan bool {
	subscriber := make(chan bool)
	c.clientDisconnectSubscribers = append(c.clientDisconnectSubscribers, subscriber)
	return subscriber
}

func (c ClientNotifier) processMutation(mutation Mutation) {
	c.mutations <- mutation
}

func newClientNotifier(dataCollector obs.DataCollector, conn connection.Connection, clientID uint64) *ClientNotifier {
	mutations := make(chan Mutation, clientBufferSize)
	go func() {
		for mutation := range mutations {
			message := MutationMessage{
				CollectionType: mutation.CollectionType,
				MutationType:   mutation.MutationType,
				Payload:        mutation.Payload,
			}
			jsonBuf, err := json.MarshalIndent(message, "", "  ")
			if err != nil {
				dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
				continue
			}
			conn.SendMessage(jsonBuf)
			dataCollector.Logger.Log(obs.Info, obs.Props{
				obs.MessageProp: obs.Props{
					"summary":    "notification sent",
					"clientID":   clientID,
					"mutationID": mutation.ID,
				},
			})
		}
	}()
	clientNotifier := &ClientNotifier{
		clientDisconnectSubscribers: make([]chan bool, 0),
		mutations:                   mutations,
	}
	go func() {
		<-conn.OnClientDisconnect()
		for _, subscriber := range clientNotifier.clientDisconnectSubscribers {
			subscriber <- true
		}
	}()
	return clientNotifier
}
