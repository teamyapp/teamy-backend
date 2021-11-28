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

// In Memory Database
type Data struct {
	lock           *sync.Mutex
	file           string
	Tasks          map[graphql.ID]Task
	Users          map[graphql.ID]User
	LifetimeEvents []LifetimeEvent
}

func (d Data) GetTask(id graphql.ID) (Task, error) {
	task, ok := d.Tasks[id]
	if ok {
		return task, nil
	}
	return task, fmt.Errorf("can not find task %v", id)
}

func (d Data) GetTasks(ids []graphql.ID) []Task {
	var tasks []Task
	for _, id := range ids {
		task, ok := d.Tasks[id]
		if ok {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (d Data) FilterTasks(filter func(t Task) bool) []Task {
	var tasks []Task
	for _, task := range d.Tasks {
		if filter(task) {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (d Data) CreateTask(task TaskInput, creatorID graphql.ID) (Task, error) {
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
	return d.Tasks[newID], d.Write()
}

func (d Data) GetUser(id graphql.ID) (User, error) {
	user, ok := d.Users[id]
	if !ok {
		return User{}, fmt.Errorf("user %v not found", id)
	}
	return user, nil
}

func (d *Data) FilterLifetimeEvents(filter func(LifetimeEvent) bool) []LifetimeEvent {
	var events []LifetimeEvent
	fmt.Println(&d, "FilterLifetimeEvents", d.LifetimeEvents)
	for _, e := range d.LifetimeEvents {
		if filter(e) {
			events = append(events, e)
		}
	}
	return events
}

func (d *Data) CreateLifetimeEvent(creatorID graphql.ID, eventType LifetimeEventType) error {
	d.LifetimeEvents = append(d.LifetimeEvents, LifetimeEvent{
		ID:         graphql.ID(strconv.FormatInt(int64(len(d.LifetimeEvents)), 10)),
		ActorID:    creatorID,
		HappensAt_: time.Now(),
		EventType:  eventType,
	})
	fmt.Println(&d, "CreateLifetimeEvent", d.LifetimeEvents)
	return d.Write()
}

func (d Data) Write() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.file == "" {
		log.Println("this data object has no persisted layer, skip file writes")
		return nil
	}

	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(d.file, bytes, os.ModePerm)
}

func Read(path string) *Data {
	data := Data{}
	defer func() {
		data.lock = &sync.Mutex{}
		if data.Tasks == nil {
			data.Tasks = make(map[graphql.ID]Task)
		}
		if data.Users == nil {
			data.Users = make(map[graphql.ID]User)
		}
	}()
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Println(err, ", fail to load data from json, skip persistence")
		return &data
	}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		log.Println(err, "fail to load data from json, skip persistence")
		return &data
	}
	data.file = path
	return &data
}
