package datastore

import (
	"fmt"
	"strconv"
	"sync"
	"time"

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
		ds.data.Users = make(map[graphql.ID]entity.User)
	}
	// todo: implement a GraphQL to create Teams
	ds.data.Teams = append(ds.data.Teams, entity.Team{
		Entity: oneEntity.Entity{
			ID: 1,
		},
	})
	return &ds
}

//
// User
//
func (d DataStore) GetUser(id graphql.ID) (entity.User, error) {
	user, ok := d.data.Users[id]
	if !ok {
		return entity.User{}, fmt.Errorf("user %v is not found", id)
	}
	return user, nil
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

func (d DataStore) CreateTask(creatorID graphql.ID, teamID oneEntity.ID, task entity.Task) (entity.Task, error) {
	// get next id
	// todo: consider global unique id for all resources/entities
	task.CreatorID = creatorID
	idInt := len(d.data.Tasks)
	task.ID = oneEntity.ID(idInt)
	taskID := graphql.ID(fmt.Sprint(idInt))
	for {
		if _, ok := d.data.Tasks[taskID]; ok {
			idInt += 1
			task.ID = oneEntity.ID(idInt)
			taskID = graphql.ID(fmt.Sprint(idInt))
		} else {
			break
		}
	}

	d.data.Tasks[taskID] = task

	// creator
	d.data.CreationRelations = append(d.data.CreationRelations, CreationRelation{
		TaskID: taskID,
		UserID: creatorID,
	})

	// add task to team
	err := d.UpdateTeam(teamID, func(t entity.Team) entity.Team {
		t.Tasks = append(t.Tasks, task.ID)
		return t
	})
	if err != nil {
		return entity.Task{}, err
	}

	err = d.persister.Write(d.data)
	if err != nil {
		return entity.Task{}, err
	}
	return task, nil
}

func (d DataStore) UpdateTask(task entity.Task) (entity.Task, error) {
	id := graphql.ID(fmt.Sprintf("%v", task.ID))
	task, ok := d.data.Tasks[id]
	if !ok {
		return entity.Task{}, fmt.Errorf("task not found: id=%v", task.ID)
	}
	d.data.Tasks[id] = task
	return task, d.persister.Write(d.data)
}

//
// Comment
//
func (d DataStore) CreateComment(comment entity.Comment) (entity.Comment, error) {
	comment.ID = graphql.ID(fmt.Sprint(len(d.data.Comments)))
	d.data.Comments = append(d.data.Comments, comment)
	return comment, d.persister.Write(d.data)
}

func (d DataStore) FilterComments(filter func(entity.Comment) bool) (cs []entity.Comment) {
	for _, c := range d.data.Comments {
		if filter(c) {
			cs = append(cs, c)
		}
	}
	return cs
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

//
// Team
//
func (d DataStore) UpdateTeam(teamID oneEntity.ID, apply func(entity.Team) entity.Team) error {
	for i, team := range d.data.Teams {
		if team.ID == teamID {
			d.data.Teams[i] = apply(team)
			return d.persister.Write(d.data)
		}
	}
	return fmt.Errorf("team %v is not found", teamID)
}

func (d DataStore) CreateTeam(t entity.Team) error {
	t.ID = oneEntity.ID(len(d.data.Teams) + 1)
	d.data.Teams = append(d.data.Teams, t)
	return d.persister.Write(d.data)
}

func (d DataStore) FilterTeams(filter func(entity.Team) bool) (ts []entity.Team) {
	for _, t := range d.data.Teams {
		if filter(t) {
			ts = append(ts, t)
		}
	}
	return
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

func toEntityID(id graphql.ID) (int, error) {
	i, err := strconv.ParseInt(string(id), 10, 32)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}
