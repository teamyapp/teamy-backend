package gqlv2

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

//
// In Memory Database
type Data struct {
	lock           *sync.Mutex
	file           string
	Tasks          map[int32]Task
	Users          map[int32]User
	LifetimeEvents []LifetimeEvent
}

func (d Data) GetTask(id int32) (Task, error) {
	task, ok := d.Tasks[id]
	if ok {
		return task, nil
	}
	return task, fmt.Errorf("can not find task %v", id)
}

func (d Data) GetTasks(ids []int32) []Task {
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

func (d Data) CreateTask(task TaskInput, creatorID int32) (Task, error) {
	newID := int32(len(d.Tasks)) + 1
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
		ID:         int32(len(d.LifetimeEvents)),
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

func (d Data) GetUser(id int32) (User, error) {
	user, ok := d.Users[id]
	if !ok {
		return User{}, fmt.Errorf("user %v not found", id)
	}
	return user, nil
}

func (d Data) FilterLifetimeEvents(filter func(LifetimeEvent) bool) []LifetimeEvent {
	var events []LifetimeEvent
	for _, e := range d.LifetimeEvents {
		if filter(e) {
			events = append(events, e)
		}
	}
	return events
}

func (d Data) Write() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(d.file, bytes, os.ModePerm)
}

func Read(path string) *Data {
	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	data := Data{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}
	data.file = path
	data.lock = &sync.Mutex{}
	return &data
}
