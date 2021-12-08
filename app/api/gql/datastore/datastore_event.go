package datastore

import (
	"time"

	oneEntity "github.com/teamyapp/one/entity"
)

func (d *DataStore) CreateLifetimeEvent(creatorID oneEntity.ID, eventType LifetimeEventType) error {
	d.data.LifetimeEvents = append(d.data.LifetimeEvents, LifetimeEvent{
		ID:         d.newID(Event),
		ActorID:    creatorID,
		HappensAt_: time.Now(),
		EventType:  eventType,
	})
	return d.persister.Write(d.data)
}

func (d *DataStore) FilterLifetimeEvents(filter func(LifetimeEvent) bool) []LifetimeEvent {
	var events []LifetimeEvent
	for _, e := range d.data.LifetimeEvents {
		if filter(e) {
			events = append(events, e)
		}
	}
	return events
}
