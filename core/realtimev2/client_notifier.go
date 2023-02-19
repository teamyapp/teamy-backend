package realtimev2

import (
	"context"
	"encoding/json"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const clientBufferSize = 50

type ClientNotifier struct {
	dataCollector               telemetry.DataCollector
	clientDisconnectSubscribers []chan bool
	clientID                    uint64
	messages                    chan Message
	acceptTransaction           bool
}

func (c *ClientNotifier) getClientID() uint64 {
	return c.clientID
}

func (c *ClientNotifier) subscribeClientDisconnect() <-chan bool {
	subscriber := make(chan bool)
	c.clientDisconnectSubscribers = append(c.clientDisconnectSubscribers, subscriber)
	return subscriber
}

func (c *ClientNotifier) onInitialStateReady() {
	c.acceptTransaction = true
}

func (c *ClientNotifier) notifyTransaction(ct context.Context, clientTransaction *ClientTransaction) {
	ct = ctx.WithClientID(ct, c.clientID)
	c.dataCollector.Logger.InfoWithContext(ct, "process transaction")

	if !c.acceptTransaction {
		c.dataCollector.Logger.InfoWithContext(ct, "discard transaction")
		return
	}

	message := Message{
		Type:    TransactionMessageType,
		Payload: clientTransaction.ToMessage(),
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

func newClientNotifier(dataCollector telemetry.DataCollector, conn connection.Connection, clientID uint64) *ClientNotifier {
	messages := make(chan Message, clientBufferSize)
	ct := context.Background()
	ct = ctx.WithClientID(ct, clientID)
	go func() {
		for message := range messages {
			jsonBuf, err := json.MarshalIndent(message, "", "  ")
			if err != nil {
				dataCollector.Logger.ErrorWithContext(ct, &errs.Error{
					Code:     errs.Serialization,
					EmbedErr: err,
				})
				continue
			}

			conn.SendMessage(jsonBuf)
			dataCollector.Logger.InfoWithContext(ct, "notification sent")
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
