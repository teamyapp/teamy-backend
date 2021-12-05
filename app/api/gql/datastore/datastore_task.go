package datastore

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

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
	task.ID = d.newID(Task)
	task.CreatorID = creatorID
	taskID := graphql.ID(fmt.Sprint(task.ID))
	d.data.Tasks[taskID] = task
	// creator
	d.data.CreationRelations = append(d.data.CreationRelations, CreationRelation{
		TaskID: taskID,
		UserID: creatorID,
	})

	err := d.persister.Write(d.data)
	if err != nil {
		return entity.Task{}, err
	}
	return task, nil
}

func (d DataStore) UpdateTask(task entity.Task) (entity.Task, error) {
	id := graphql.ID(fmt.Sprintf("%v", task.ID))
	if _, ok := d.data.Tasks[id]; !ok {
		return entity.Task{}, fmt.Errorf("task not found: id=%v", task.ID)
	}
	d.data.Tasks[id] = task
	return task, d.persister.Write(d.data)
}

func (d DataStore) DeleteTask(taskID graphql.ID) (entity.Task, error) {
	task, ok := d.data.Tasks[taskID]
	if !ok {
		return entity.Task{}, errors.Errorf("task %v does not exist", taskID)
	}
	delete(d.data.Tasks, taskID)
	return task, nil
}
