package datastore

import (
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
