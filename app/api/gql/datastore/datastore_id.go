package datastore

func (d *DataStore) newID(entityType Type) uint64 {
	newID := uint64(len(d.data.IDs) + 1)
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
