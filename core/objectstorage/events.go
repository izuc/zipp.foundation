package objectstorage

import (
	"github.com/izuc/zipp.foundation/core/events"
)

type Events struct {
	ObjectEvicted *events.Event
}

func evictionEvent(handler interface{}, params ...interface{}) {
	handler.(func([]byte, StorableObject))(params[0].([]byte), params[1].(StorableObject))
}
