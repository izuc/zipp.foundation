//go:build zippdb

package zippdb

import (
	"fmt"

	"github.com/izuc/zipp.foundation/runtime/ioutils"
	"github.com/izuc/zippdb"
)

// ZIPPDB holds the underlying zippdb.DB instance and options.
type ZIPPDB struct {
	db *zippdb.DB
	ro *zippdb.ReadOptions
	wo *zippdb.WriteOptions
	fo *zippdb.FlushOptions
}

// CreateDB creates a new ZIPPDB instance.
func CreateDB(directory string, options ...Option) (*ZIPPDB, error) {

	if err := ioutils.CreateDirectory(directory, 0700); err != nil {
		return nil, fmt.Errorf("could not create directory: %w", err)
	}

	dbOpts := dbOptions(options)

	opts := zippdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	opts.SetCompression(zippdb.NoCompression)
	if dbOpts.compression {
		opts.SetCompression(zippdb.ZSTDCompression)
	}

	if dbOpts.parallelism > 0 {
		opts.IncreaseParallelism(dbOpts.parallelism)
	}

	for _, str := range dbOpts.custom {
		var err error
		opts, err = zippdb.GetOptionsFromString(opts, str)
		if err != nil {
			return nil, err
		}
	}

	ro := zippdb.NewDefaultReadOptions()
	ro.SetFillCache(dbOpts.fillCache)

	wo := zippdb.NewDefaultWriteOptions()
	wo.SetSync(dbOpts.sync)
	wo.DisableWAL(dbOpts.disableWAL)

	fo := zippdb.NewDefaultFlushOptions()

	db, err := zippdb.OpenDb(opts, directory)
	if err != nil {
		return nil, err
	}

	return &ZIPPDB{
		db: db,
		ro: ro,
		wo: wo,
		fo: fo,
	}, nil
}

// OpenDBReadOnly opens a new ZIPPDB instance in read-only mode.
func OpenDBReadOnly(directory string, options ...Option) (*ZIPPDB, error) {

	dbOpts := dbOptions(options)

	opts := zippdb.NewDefaultOptions()
	opts.SetCompression(zippdb.NoCompression)
	if dbOpts.compression {
		opts.SetCompression(zippdb.ZSTDCompression)
	}

	for _, str := range dbOpts.custom {
		var err error
		opts, err = zippdb.GetOptionsFromString(opts, str)
		if err != nil {
			return nil, err
		}
	}

	ro := zippdb.NewDefaultReadOptions()
	ro.SetFillCache(dbOpts.fillCache)

	db, err := zippdb.OpenDbForReadOnly(opts, directory, true)
	if err != nil {
		return nil, err
	}

	return &ZIPPDB{
		db: db,
		ro: ro,
	}, nil
}

func dbOptions(optionalOptions []Option) *Options {
	result := &Options{
		compression: false,
		fillCache:   false,
		sync:        false,
		disableWAL:  true,
		parallelism: 0,
	}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}
	return result
}

// Flush the database.
func (r *ZIPPDB) Flush() error {
	return r.db.Flush(r.fo)
}

// Close the database.
func (r *ZIPPDB) Close() error {
	r.db.Close()
	return nil
}

// GetProperty returns the value of a database property.
func (r *ZIPPDB) GetProperty(name string) string {
	return r.db.GetProperty(name)
}

// GetIntProperty similar to "GetProperty", but only works for a subset of properties whose
// return value is an integer. Return the value by integer.
func (r *ZIPPDB) GetIntProperty(name string) (uint64, bool) {
	return r.db.GetIntProperty(name)
}
