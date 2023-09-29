//go:build !zippdb

package zippdb

import "github.com/izuc/zipp.foundation/kvstore"

const (
	panicMissingZIPPDB = "For ZIPPDB support please compile with '-tags zippdb'"
)

// ZIPPDB holds the underlying zippdb.DB instance and options.
type ZIPPDB struct {
}

// CreateDB creates a new ZIPPDB instance.
func CreateDB(directory string, options ...Option) (*ZIPPDB, error) {
	panic(panicMissingZIPPDB)
}

// OpenDBReadOnly opens a new ZIPPDB instance in read-only mode.
func OpenDBReadOnly(directory string, options ...Option) (*ZIPPDB, error) {
	panic(panicMissingZIPPDB)
}

// New creates a new KVStore with the underlying ZIPPDB.
func New(db *ZIPPDB) kvstore.KVStore {
	panic(panicMissingZIPPDB)
}

// Flush the database.
func (r *ZIPPDB) Flush() error {
	panic(panicMissingZIPPDB)
}

// Close the database.
func (r *ZIPPDB) Close() error {
	panic(panicMissingZIPPDB)
}

// GetProperty returns the value of a database property.
func (r *ZIPPDB) GetProperty(name string) string {
	panic(panicMissingZIPPDB)
}

// GetIntProperty similar to "GetProperty", but only works for a subset of properties whose
// return value is an integer. Return the value by integer.
func (r *ZIPPDB) GetIntProperty(name string) (uint64, bool) {
	panic(panicMissingZIPPDB)
}
