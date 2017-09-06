package instance

import (
	"time"

	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/history"
	"github.com/weaveworks/flux/service"
)

type EventReadWriter struct {
	inst service.InstanceID
	db   history.DB
}

func (rw EventReadWriter) LogEvent(e history.Event) error {
	return rw.db.LogEvent(rw.inst, e)
}

func (rw EventReadWriter) AllEvents(before time.Time, limit int64, after time.Time) ([]history.Event, error) {
	return rw.db.AllEvents(rw.inst, before, limit, after)
}

func (rw EventReadWriter) EventsForService(service flux.ResourceID, before time.Time, limit int64, after time.Time) ([]history.Event, error) {
	return rw.db.EventsForService(rw.inst, service, before, limit, after)
}

func (rw EventReadWriter) GetEvent(id history.EventID) (history.Event, error) {
	return rw.db.GetEvent(id)
}
