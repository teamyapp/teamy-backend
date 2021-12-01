package resolver

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
)

// Temperary SQL like struct for v2 migration purpose.
type CreationRelation struct {
	TaskID graphql.ID
	UserID graphql.ID
}

// In Memory Database
type DataStore struct {
	persister Persister
	lock      *sync.Mutex
	Data
}

type Data struct {
	Tasks             map[graphql.ID]Task
	Users             map[graphql.ID]User
	LifetimeEvents    []LifetimeEvent
	CreationRelations []CreationRelation
}

func NewDataStore(p Persister) *DataStore {
	ds := DataStore{
		Data:      *p.Read(),
		persister: p,
	}
	ds.lock = &sync.Mutex{}
	if ds.Tasks == nil {
		ds.Tasks = make(map[graphql.ID]Task)
	}
	if ds.Users == nil {
		ds.Users = make(map[graphql.ID]User)
	}
	return &ds
}

// Persister persists a Data instance
type Persister interface {
	Write(*Data) error
	Read() *Data
}

func (d DataStore) GetTask(id graphql.ID) (Task, error) {
	task, ok := d.Tasks[id]
	if ok {
		return task, nil
	}
	return task, fmt.Errorf("can not find task %v", id)
}

func (d DataStore) GetTasks(ids []graphql.ID) []Task {
	var tasks []Task
	for _, id := range ids {
		task, ok := d.Tasks[id]
		if ok {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (d DataStore) FilterTasks(filter func(t Task) bool) []Task {
	var tasks []Task
	for _, task := range d.Tasks {
		if filter(task) {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (d DataStore) CreateTask(task TaskInput, creatorID graphql.ID) (Task, error) {
	newID := graphql.ID(strconv.FormatInt(int64(len(d.Tasks))+1, 10))
	context := ""
	if task.Context != nil {
		context = *task.Context
	}
	d.Tasks[newID] = Task{
		ID:        newID,
		Goal:      task.Goal,
		Context:   context,
		CreatorID: creatorID,
	}
	d.LifetimeEvents = append(d.LifetimeEvents, LifetimeEvent{
		ID:         graphql.ID(strconv.FormatInt(int64(len(d.LifetimeEvents)), 10)),
		ActorID:    creatorID,
		HappensAt_: time.Now(),
		EventType: LifetimeEventType{
			Type: Creation,
			Creation: &EventCreation{
				TaskID: newID,
			},
		},
	})
	return d.Tasks[newID], d.persister.Write(&d.Data)
}

func (d DataStore) GetUser(id graphql.ID) (User, error) {
	user, ok := d.Users[id]
	if !ok {
		return User{}, fmt.Errorf("user %v not found", id)
	}
	return user, nil
}

func (d *DataStore) FilterLifetimeEvents(filter func(LifetimeEvent) bool) []LifetimeEvent {
	var events []LifetimeEvent
	fmt.Println(&d, "FilterLifetimeEvents", d.LifetimeEvents)

	for _, e := range d.LifetimeEvents {
		if filter(e) {
			events = append(events, e)
		}
	}
	return events
}

func (d *DataStore) CreateLifetimeEvent(creatorID graphql.ID, eventType LifetimeEventType) error {
	d.LifetimeEvents = append(d.LifetimeEvents, LifetimeEvent{
		ID:         graphql.ID(strconv.FormatInt(int64(len(d.LifetimeEvents)), 10)),
		ActorID:    creatorID,
		HappensAt_: time.Now(),
		EventType:  eventType,
	})
	fmt.Println(&d, "CreateLifetimeEvent", d.LifetimeEvents)
	return d.persister.Write(&d.Data)
}

func (d *DataStore) FilterCreationRelation(f func(CreationRelation) bool) (rs []CreationRelation) {
	for _, r := range d.CreationRelations {
		if f(r) {
			rs = append(rs, r)
		}
	}
	return
}

/////////////////
// Persistance //
/////////////////
type JSONPersister struct {
	File string
}

var _ Persister = (*JSONPersister)(nil)

func NewJSONPersister() JSONPersister {
	return JSONPersister{
		File: "./data.json",
	}
}

func (p JSONPersister) Write(d *Data) error {
	if p.File == "" {
		log.Println("this data object has no persisted layer, skip file writes")
		return nil
	}

	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.File, bytes, os.ModePerm)
}

func (p JSONPersister) Read() *Data {
	data := &Data{}
	bytes, err := os.ReadFile(p.File)
	if err != nil {
		log.Println(err, ", fail to load data from json, skip persistence")
		return data
	}
	err = json.Unmarshal(bytes, data)
	if err != nil {
		log.Println(err, "fail to load data from json, skip persistence")
		return data
	}
	return data
}
