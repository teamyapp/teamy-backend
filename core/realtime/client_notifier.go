package realtime

import (
	"encoding/json"
	"log"

	"github.com/teamyapp/cloud/libs/connection"
)

const clientBufferSize = 50

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
			jsonBuf, err := json.MarshalIndent(mutation, "", "  ")
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
