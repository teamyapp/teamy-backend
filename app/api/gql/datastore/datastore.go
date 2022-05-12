package datastore

import (
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

// Persister persists a Data instance
type Persister interface {
	Write(*Data) error
	Read() *Data
}

// In Memory Database
type DataStore struct {
	persister Persister
	lock      *sync.Mutex
	data      *Data
}

func NewDataStore(p Persister) *DataStore {
	ds := DataStore{
		data:      p.Read(),
		persister: p,
	}
	ds.lock = &sync.Mutex{}
	if ds.data.Tasks == nil {
		ds.data.Tasks = make(map[graphql.ID]entity.Task)
	}
	if ds.data.Users == nil {
		ds.data.Users = make(map[uint64]entity.User)
	}
	if ds.data.IDs == nil {
		ds.data.IDs = make(map[uint64]Type)
	}
	if ds.data.Comments == nil {
		ds.data.Comments = make(map[uint64]entity.Comment)
	}
	if ds.data.Invitations == nil {
		ds.data.Invitations = make(map[uint64]entity.Invitation)
	}
	for i, team := range ds.data.Teams {
		// maintain the set
		var members entity.OrderedSetID
		for _, member := range team.MemberIDs {
			members = members.Add(member)
		}
		ds.data.Teams[i].MemberIDs = members
		var tasks entity.OrderedSetID
		for _, taskID := range team.Tasks {
			tasks = tasks.Add(taskID)
		}
		ds.data.Teams[i].Tasks = tasks
	}
	return &ds
}

func (d *DataStore) FilterCreationRelation(filter func(CreationRelation) bool) (rs []CreationRelation) {
	// fmt.Printf("%+v", d.data.CreationRelations)
	for _, r := range d.data.CreationRelations {
		if filter(r) {
			rs = append(rs, r)
		}
	}
	return
}
