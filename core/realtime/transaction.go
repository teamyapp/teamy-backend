package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type ClientTransaction struct {
	logger         telemetry.Logger
	id             uint64
	mutations      []Mutation
	clientNotifier *ClientNotifier
}

func (c *ClientTransaction) addMutation(mutation Mutation) {
	c.mutations = append(c.mutations, mutation)
}

// Deprecated: The old method should not be used anymore. Use notify instead
func (c *ClientTransaction) commit(ct context.Context) {
	c.clientNotifier.notifyTransaction(ct, c)
}

func (c *ClientTransaction) notify(ct context.Context) {
	c.clientNotifier.notifyTransaction(ct, c)
}

func (c *ClientTransaction) ToMessage() TransactionMessage {
	mutationMessages := collect.Map(c.mutations, func(mutation Mutation, _ int) MutationMessage {
		return mutation.ToMessage()
	})
	return TransactionMessage{
		ID:        c.id,
		Mutations: mutationMessages,
	}
}

func newClientTransaction(logger telemetry.Logger, clientNotifier *ClientNotifier, id uint64) *ClientTransaction {
	return &ClientTransaction{
		logger:         logger,
		clientNotifier: clientNotifier,
		id:             id,
	}
}

type Transaction struct {
	logger            telemetry.Logger
	stateSyncer       *StateSyncer
	id                uint64
	mutations         []Mutation
	nextMutationIndex int
}

func (t *Transaction) rollback(ct context.Context) *errs.Error {
	for index := t.nextMutationIndex - 1; index >= 0; index-- {
		err := t.mutations[index].Undo()
		if err != nil {
			t.logger.ErrorWithContext(ct, err)
		}
	}

	return nil
}

func (t *Transaction) AppendMutation(mutation Mutation) {
	t.mutations = append(t.mutations, mutation)
}

func (t *Transaction) GetMutations() []Mutation {
	return t.mutations
}

func (t *Transaction) Notify(ct context.Context) *errs.Error {
	t.stateSyncer.BeginTransaction()
	defer t.stateSyncer.EndTransaction()

	clientTransactions := make(map[uint64]*ClientTransaction)
	for _, mutation := range t.mutations {
		clientNotifiers := mutation.GetClientNotifiers()
		for _, clientNotifier := range clientNotifiers {
			clientID := clientNotifier.getClientID()
			_, ok := clientTransactions[clientID]
			if !ok {
				clientTransactionID := t.stateSyncer.NextClientTransactionID()
				clientTransactions[clientID] = newClientTransaction(
					clientNotifier.logger,
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
			return err
		}
	}

	return nil
}

func NewTransaction(logger telemetry.Logger, stateSyncer *StateSyncer) *Transaction {
	return &Transaction{
		logger:            logger,
		stateSyncer:       stateSyncer,
		id:                stateSyncer.NextTransactionID(),
		mutations:         make([]Mutation, 0),
		nextMutationIndex: 0,
	}
}
