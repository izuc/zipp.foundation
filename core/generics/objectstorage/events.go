package objectstorage

import "github.com/izuc/zipp.foundation/core/objectstorage"

type Events = objectstorage.Events

func evictionEvent[T StorableObject](handler interface{}, params ...interface{}) {
	handler.(func([]byte, T))(params[0].([]byte), params[1].(T))
}
