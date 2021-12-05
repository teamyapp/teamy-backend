package datastore

import (
	"strconv"
	"sync"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
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
		ds.data.Users = make(map[oneEntity.ID]entity.User)
	}
	if ds.data.IDs == nil {
		ds.data.IDs = make(map[oneEntity.ID]Type)
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

func toEntityID(id graphql.ID) (oneEntity.ID, error) {
	i, err := strconv.ParseInt(string(id), 10, 32)
	if err != nil {
		return 0, err
	}
	return oneEntity.ID(i), nil
}
