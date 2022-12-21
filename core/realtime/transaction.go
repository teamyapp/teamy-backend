package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
)

type MutationInput struct {
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
}

func (m *MutationInput) toMutation(id uint64) Mutation {
	return Mutation{
		ID:             id,
		CollectionType: m.CollectionType,
		MutationType:   m.MutationType,
		Payload:        m.Payload,
	}
}

// Transaction contains all the mutations for a single team
type Transaction struct {
	dataCollector obs.DataCollector
	stateSyncer   *StateSyncer
	id            uint64
	mutations     []Mutation
	teamID        uint64
}

func (t *Transaction) AddMutation(ct context.Context, input MutationInput) {
	mutationID := t.stateSyncer.NextMutationID()
	mutation := input.toMutation(mutationID)
	t.mutations = append(t.mutations, mutation)
}

func NewTransaction(
	stateSyncer *StateSyncer,
	dataCollector obs.DataCollector,
	teamID uint64) *Transaction {
	return &Transaction{
		id:          stateSyncer.NextTransactionID(),
		stateSyncer: stateSyncer,
		mutations:   make([]Mutation, 0),
		teamID:      teamID,
	}
}
