package datastore

import (
	"fmt"
	"strconv"
	"sync"
	"time"

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
		ds.data.Users = make(map[graphql.ID]entity.User)
	}
	return &ds
}

//
// Tasks
//
func (d DataStore) GetTask(id graphql.ID) (entity.Task, error) {
	task, ok := d.data.Tasks[id]
	if ok {
		return task, nil
	}
	return task, fmt.Errorf("can not find task %v", id)
}

func (d DataStore) GetTasks(ids []graphql.ID) []entity.Task {
	var tasks []entity.Task
	for _, id := range ids {
		task, ok := d.data.Tasks[id]
		if ok {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (d DataStore) FilterTasks(filter func(t entity.Task) bool) []entity.Task {
	var tasks []entity.Task
	for _, task := range d.data.Tasks {
		if filter(task) {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (d DataStore) CreateTask(creatorID graphql.ID, task entity.Task) entity.Task {
	id := graphql.ID(fmt.Sprintf("%v", task.ID))
	d.data.Tasks[id] = task
	d.data.CreationRelations = append(d.data.CreationRelations, CreationRelation{
		TaskID: id,
		UserID: creatorID,
	})
	return task
}

func (d DataStore) UpdateTask(task entity.Task) (entity.Task, error) {
	id := graphql.ID(fmt.Sprintf("%v", task.ID))
	d.data.Tasks[id] = task
	return task, d.persister.Write(d.data)
}

//
// Lifetime Events
//
func (d *DataStore) FilterLifetimeEvents(filter func(LifetimeEvent) bool) []LifetimeEvent {
	var events []LifetimeEvent
	for _, e := range d.data.LifetimeEvents {
		if filter(e) {
			events = append(events, e)
		}
	}
	return events
}

func (d *DataStore) CreateLifetimeEvent(creatorID graphql.ID, eventType LifetimeEventType) error {
	d.data.LifetimeEvents = append(d.data.LifetimeEvents, LifetimeEvent{
		ID:         graphql.ID(strconv.FormatInt(int64(len(d.data.LifetimeEvents)), 10)),
		ActorID:    creatorID,
		HappensAt_: time.Now(),
		EventType:  eventType,
	})
	return d.persister.Write(d.data)
}

func (d *DataStore) FilterCreationRelation(f func(CreationRelation) bool) (rs []CreationRelation) {
	// fmt.Printf("%+v", d.data.CreationRelations)
	for _, r := range d.data.CreationRelations {
		if f(r) {
			rs = append(rs, r)
		}
	}
	return
}
