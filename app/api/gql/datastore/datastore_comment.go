package datastore

import (
	"fmt"
	"time"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

func (d DataStore) CreateComment(comment entity.Comment) (entity.Comment, error) {
	comment.ID = d.newID(Comment)
	comment.CreatedAt = time.Now()
	d.data.Comments[comment.ID] = comment
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

func (d DataStore) UpdateComment(id oneEntity.ID, apply func(entity.Comment) entity.Comment) (entity.Comment, error) {
	comment, ok := d.data.Comments[id]
	if ok {
		d.data.Comments[id] = apply(comment)
		return d.data.Comments[id], d.persister.Write(d.data)
	}
	return entity.Comment{}, fmt.Errorf("comment %v is not found", id)
}

func (d DataStore) DeleteComment(id oneEntity.ID) (entity.Comment, error) {
	comment, ok := d.data.Comments[id]
	if ok {
		delete(d.data.Comments, id)
		return comment, d.persister.Write(d.data)
	}
	return entity.Comment{}, fmt.Errorf("comment %v is not found", id)
}
