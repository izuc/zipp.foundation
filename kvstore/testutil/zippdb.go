package testutil

import (
	"strconv"
	"testing"

	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/kvstore/zippdb"
)

// ZIPPDB creates a temporary ZIPPDBKVStore that automatically gets cleaned up when the test finishes.
func ZIPPDB(t *testing.T) (kvstore.KVStore, error) {
	dir := t.TempDir()

	db, err := zippdb.CreateDB(dir)
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Errorf("Closing database: %v", err)
		}
	})

	databaseCounterMutex.Lock()
	databaseCounter[t.Name()]++
	counter := databaseCounter[t.Name()]
	databaseCounterMutex.Unlock()

	storeWithRealm, err := zippdb.New(db).WithRealm([]byte(t.Name() + strconv.Itoa(counter)))
	if err != nil {
		return nil, err
	}

	return storeWithRealm, nil
}
