package datastore

import (
	"fmt"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

func (d DataStore) CreateComment(comment entity.Comment) (entity.Comment, error) {
	comment.ID = d.newID(Comment)
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

func (d DataStore) UpdateComment(id oneEntity.ID, apply func(entity.Comment) entity.Comment) (entity.Comment, error) {
	for i, comment := range d.data.Comments {
		if comment.ID == id {
			newC := apply(comment)
			d.data.Comments[i] = newC
			return newC, d.persister.Write(d.data)
		}
	}
	return entity.Comment{}, fmt.Errorf("comment %v is not found", id)
}
