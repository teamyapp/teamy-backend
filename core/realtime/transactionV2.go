package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type ClientTransactionV2 struct {
	dataCollector  telemetry.DataCollector
	id             uint64
	mutations      []Mutation
	clientNotifier *ClientNotifier
}

func (c *ClientTransactionV2) addMutation(mutation Mutation) {
	c.mutations = append(c.mutations, mutation)
}

func (c *ClientTransactionV2) notify(ct context.Context) {
	c.clientNotifier.notifyTransactionV2(ct, c)
}

func (c *ClientTransactionV2) ToMessage() TransactionMessage {
	mutationMessages := collect.Map(c.mutations, func(mutation Mutation, _ int) MutationMessage {
		return mutation.ToMessage()
	})
	return TransactionMessage{
		ID:        c.id,
		Mutations: mutationMessages,
	}
}

func newClientTransactionV2(dataCollector telemetry.DataCollector, clientNotifier *ClientNotifier, id uint64) *ClientTransactionV2 {
	return &ClientTransactionV2{
		dataCollector:  dataCollector,
		clientNotifier: clientNotifier,
		id:             id,
	}
}

type TransactionV2 struct {
	dataCollector     telemetry.DataCollector
	stateSyncer       *StateSyncer
	id                uint64
	mutations         []Mutation
	nextMutationIndex int
}

func (t *TransactionV2) AppendMutation(ct context.Context, mutation Mutation) *errs.Error {
	t.mutations = append(t.mutations, mutation)
	return nil
}

func (t *TransactionV2) Notify(ct context.Context) *errs.Error {
	t.stateSyncer.BeginTransaction()
	defer t.stateSyncer.EndTransaction()

	clientTransactions := make(map[uint64]*ClientTransactionV2)
	for _, mutation := range t.mutations {
		clientNotifiers, err := mutation.GetClientNotifiers(ct)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		for _, clientNotifier := range clientNotifiers {
			clientID := clientNotifier.getClientID()
			_, ok := clientTransactions[clientID]
			if !ok {
				clientTransactionID := t.stateSyncer.NextClientTransactionID()
				clientTransactions[clientID] = newClientTransactionV2(
					clientNotifier.dataCollector,
					clientNotifier,
					clientTransactionID,
				)
			}

			clientTransaction := clientTransactions[clientID]
			clientTransaction.addMutation(mutation)
		}
	}

	for _, clientTransaction := range clientTransactions {
		clientTransaction.notify(ct)
	}

	for _, mutation := range t.mutations {
		err := mutation.CleanUp(ct)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}
	}

	return nil
}

func NewTransactionV2(dataCollector telemetry.DataCollector, stateSyncer *StateSyncer) *TransactionV2 {
	return &TransactionV2{
		dataCollector:     dataCollector,
		stateSyncer:       stateSyncer,
		id:                stateSyncer.NextTransactionID(),
		mutations:         make([]Mutation, 0),
		nextMutationIndex: 0,
	}
}
