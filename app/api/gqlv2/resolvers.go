package gqlv2

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed graphiql.html
var graphiql []byte

func QraphiQL() []byte {
	return graphiql
}

//go:embed schema.gql
var rawSchema string

func RawSchema() string {
	return rawSchema
}

type Dependencies struct {
	Data *Data
}

type Root struct {
	Deps Dependencies
}

func (r Root) Tasks(args struct{ ID int32 }) ([]Task, error) {
	tasks := r.Deps.Data.GetTasks([]int32{args.ID})
	for i := range tasks {
		tasks[i].deps = r.Deps
	}
	return tasks, nil
}

func (r Root) Me() User {
	return User{}
}

type TaskInput struct {
	Goal    string
	Context *string
}

func (r Root) CreateTask(args struct{ Input TaskInput }) (Task, error) {
	return r.Deps.Data.CreateTask(args.Input, 1)
}

type Task struct {
	// this is always injected by a parent level resolver which returns a Task, at runtime
	deps Dependencies
	// these 4 can be fetched by a parent level resolver which returns a Task
	// in terms of SQL, they could be part of the Task table
	// other method resolvers could be "inner joins"
	ID      int32
	Goal    string
	Context string
	DueAt   *time.Time

	// foreign keys
	CreatorID int32
}

// Mentioned could be a function of Goal and Context
func (t Task) Mentioned() []Mentionable {
	parseMentioned := func(input string) (m []Mentionable) {
		chunks := strings.Split(input, " ")
		for _, chunk := range chunks {
			if len(chunk) == 0 {
				continue
			}
			if chunk[0] == '@' {
				id, err := strconv.ParseInt(chunk[1:], 10, 32)
				if err != nil {
					continue
				}
				m = append(m, Mentionable{
					dep:  t.deps,
					Type: "User",
					ID:   int32(id),
				})
			} else if chunk[0] == '#' {
				id, err := strconv.ParseInt(chunk[1:], 10, 32)
				if err != nil {
					continue
				}
				m = append(m, Mentionable{
					dep:  t.deps,
					Type: "Task",
					ID:   int32(id),
				})
			}
		}
		return
	}
	return parseMentioned(t.Context)
}
func (t Task) Comments() []Comment { return []Comment{} }
func (t Task) DependsOn() []Task   { return []Task{} }
func (t Task) Creator() (User, error) {
	user, err := t.deps.Data.GetUser(t.CreatorID)
	user.dep = t.deps
	return user, err
}
func (t Task) Assignees() []User                   { return []User{} }
func (t Task) LifetimeEvents() []TaskLifetimeEvent { return []TaskLifetimeEvent{} }

type Mentionable struct {
	dep  Dependencies
	Type string
	ID   int32
}

func (m Mentionable) ToUser() (*User, bool) {
	u, err := m.dep.Data.GetUser(m.ID)
	if err != nil {
		return nil, false
	}
	return &u, true
}

func (m Mentionable) ToTask() (*Task, bool) {
	tasks := m.dep.Data.GetTasks([]int32{m.ID})
	if len(tasks) == 0 {
		return nil, false
	}
	return &tasks[0], true
}

type Comment struct{}

func (c Comment) Commenter() User {
	return User{}
}

func (c Comment) Content() string {
	return ""
}
func (t Comment) Mentioned() []Mentionable {
	return []Mentionable{}
}

type User struct {
	dep        Dependencies
	ID         int32
	Name       string
	ProfileUrl string
}

func (u User) Tasks() []Task {
	tasks := u.dep.Data.FilterTasks(func(t Task) bool {
		return t.CreatorID == u.ID
	})
	return tasks
}

func (u User) TaskNeedAttention() *Task {
	return &Task{}
}

func (u User) UpcomingTasks() []Task {
	return nil
}

func (u User) DeliveredTasks() []Task {
	return nil
}

type TaskLifetimeEvent struct{}

//
// In Memory Database
type Data struct {
	lock  *sync.Mutex
	file  string
	Tasks map[int32]Task
	Users map[int32]User
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
	return d.Tasks[newID], d.Write()
}

func (d Data) GetUser(id int32) (User, error) {
	user, ok := d.Users[id]
	if !ok {
		return User{}, fmt.Errorf("user %v not found", id)
	}
	return user, nil
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
