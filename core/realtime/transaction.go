package realtime

import (
	"context"
	"encoding/json"

	"github.com/teamyapp/cloud/libs/obs"
)

type ClientTransaction struct {
	id             uint64
	dataCollector  obs.DataCollector
	mutations      []Mutation
	clientNotifier *ClientNotifier
}

func (c *ClientTransaction) addMutation(mutation Mutation) {
	c.mutations = append(c.mutations, mutation)
}

func (c *ClientTransaction) commit(ct context.Context) {
	c.clientNotifier.notifyTransaction(ct, *c)
}

func NewClientTransaction(id uint64, dataCollector obs.DataCollector, clientNotifier *ClientNotifier) *ClientTransaction {
	return &ClientTransaction{
		id:             id,
		dataCollector:  dataCollector,
		clientNotifier: clientNotifier,
	}
}

type Transaction struct {
	id                     uint64
	dataCollector          obs.DataCollector
	mutations              []Mutation
	processedMutationIndex int
	stateSyncer            *StateSyncer
}

func (c *Transaction) logMutation(ct context.Context, mutation Mutation) {
	ct = WithMutationID(ct, mutation.GetID())
	buf, err := json.Marshal(mutation)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	} else {
		c.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
			obs.MessageProp: obs.Props{
				"Summary":  "add mutation",
				"Mutation": string(buf),
			},
		})
	}
}

func (t *Transaction) rollback(ct context.Context) error {
	for index := t.processedMutationIndex; index >= 0; index-- {
		err := t.mutations[index].Undo()
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		}
	}

	return nil
}

func (t *Transaction) AddMutation(ct context.Context, mutation Mutation) {
	t.logMutation(ct, mutation)
	t.mutations = append(t.mutations, mutation)
}

func (t *Transaction) Commit(ct context.Context) error {
	t.stateSyncer.BeginTransaction()
	defer t.stateSyncer.EndTransaction()

	clientTransactions := make(map[uint64]*ClientTransaction)
	for index, mutation := range t.mutations {
		t.processedMutationIndex = index
		err := mutation.Execute(ct)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			undoErr := t.rollback(ct)
			if undoErr != nil {
				t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: undoErr})
				return undoErr
			}

			return err
		}

		clientNotifiers, err := mutation.GetClientNotifiers(ct)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			undoErr := t.rollback(ct)
			if undoErr != nil {
				t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: undoErr})
				return undoErr
			}

			return err
		}

		for _, clientNotifier := range clientNotifiers {
			clientID := clientNotifier.getClientID()
			_, ok := clientTransactions[clientID]

			if !ok {
				clientTransactions[clientID] = NewClientTransaction(t.stateSyncer.NextClientTransactionID(), clientNotifier.dataCollector, clientNotifier)
			}

			clientTransaction := clientTransactions[clientID]
			clientTransaction.addMutation(mutation)
		}
	}

	for _, clientTransaction := range clientTransactions {
		clientTransaction.commit(ct)
	}

	return nil
}

func NewTransaction(stateSyncer *StateSyncer, dataCollector obs.DataCollector) *Transaction {
	return &Transaction{
		id:                     stateSyncer.NextTransactionID(),
		dataCollector:          dataCollector,
		mutations:              make([]Mutation, 0),
		processedMutationIndex: 0,
		stateSyncer:            stateSyncer,
	}
}
