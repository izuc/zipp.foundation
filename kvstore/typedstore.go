package kvstore

import (
	"github.com/pkg/errors"

	"github.com/izuc/zipp.foundation/serializer"
)

// TypedStore is a generically typed wrapper around a KVStore that abstracts serialization away.
type TypedStore[K, V any, KPtr serializer.MarshalablePtr[K], VPtr serializer.MarshalablePtr[V]] struct {
	kv KVStore
}

// NewTypedStore is the constructor for TypedStore.
func NewTypedStore[K, V any, KPtr serializer.MarshalablePtr[K], VPtr serializer.MarshalablePtr[V]](kv KVStore) *TypedStore[K, V, KPtr, VPtr] {
	return &TypedStore[K, V, KPtr, VPtr]{
		kv: kv,
	}
}

// Get gets the given key or an error if an error occurred.
func (t *TypedStore[K, V, KPtr, VPtr]) Get(key K) (value V, err error) {
	keyBytes, err := (KPtr)(&key).Bytes()
	if err != nil {
		return value, errors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := t.kv.Get(keyBytes)
	if err != nil {
		return value, errors.Wrap(err, "failed to retrieve from KV store")
	}

	var valuePtr VPtr = new(V)
	if _, err = valuePtr.FromBytes(valueBytes); err != nil {
		return value, errors.Wrap(err, "failed to decode value")
	}

	return *valuePtr, nil
}

// Set sets the given key and value.
func (t *TypedStore[K, V, KPtr, VPtr]) Set(key K, value V) (err error) {
	keyBytes, err := (KPtr)(&key).Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := (VPtr)(&value).Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to encode value")
	}

	err = t.kv.Set(keyBytes, valueBytes)
	if err != nil {
		return errors.Wrap(err, "failed to store in KV store")
	}

	return nil
}

// Delete deletes the given key from the store.
func (t *TypedStore[K, V, KPtr, VPtr]) Delete(key K) (err error) {
	keyBytes, err := (KPtr)(&key).Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to encode key")
	}

	err = t.kv.Delete(keyBytes)
	if err != nil {
		return errors.Wrap(err, "failed to delete entry from KV store")
	}

	return nil
}

func (t *TypedStore[K, V, KPtr, VPtr]) Iterate(prefix KeyPrefix, callback func(key K, value V) (advance bool), direction ...IterDirection) (err error) {
	return t.kv.Iterate(prefix, func(key Key, value Value) bool {
		// 1. Check for nil Values from KVStore
		if key == nil || value == nil {
			// Logging or handling the nil case, for now, we just skip this entry
			return true
		}

		// 2. Deserialize key
		var keyDecoded KPtr = new(K)
		if _, err := keyDecoded.FromBytes(key); err != nil {
			// Handle or log the deserialization error
			return false
		}

		// 2. Deserialize value
		var valueDecoded VPtr = new(V)
		if _, err := valueDecoded.FromBytes(value); err != nil {
			// Handle or log the deserialization error
			return false
		}

		// Invoke the user-provided callback
		return callback(*keyDecoded, *valueDecoded)

	}, direction...)
}
