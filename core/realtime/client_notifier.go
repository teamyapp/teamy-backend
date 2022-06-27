package realtime

import (
	"encoding/json"
	"log"

	"github.com/teamyapp/cloud/libs/connection"
)

const clientBufferSize = 50

type MutationMessage struct {
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
}

type ClientNotifier struct {
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

func newClientNotifier(conn connection.Connection) *ClientNotifier {
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
				log.Println(err)
				continue
			}

			conn.SendMessage(jsonBuf)
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
