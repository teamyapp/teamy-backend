package datastore

import oneEntity "github.com/teamyapp/one/entity"

func (d *DataStore) newID(entityType Type) oneEntity.ID {
	newID := oneEntity.ID(len(d.data.IDs) + 1)
	for {
		if _, ok := d.data.IDs[newID]; ok {
			newID += 1
		} else {
			break
		}
	}
	d.data.IDs[newID] = entityType
	return newID
}
