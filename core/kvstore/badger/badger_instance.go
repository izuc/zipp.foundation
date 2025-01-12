package badger

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"

	"github.com/izuc/zipp.foundation/core/ioutils"
)

func CreateDB(directory string, optionalOptions ...badger.Options) (*badger.DB, error) {

	if err := ioutils.CreateDirectory(directory, 0700); err != nil {
		return nil, fmt.Errorf("could not create directory: %w", err)
	}

	var opts badger.Options

	if len(optionalOptions) > 0 {
		opts = optionalOptions[0]
	} else {
		opts = badger.DefaultOptions(directory)
		opts.Logger = nil
		opts.LevelSizeMultiplier = 10
		opts.MaxLevels = 7
		opts.NumCompactors = 2 // Compactions can be expensive. Only run 2.
		opts.NumLevelZeroTables = 5
		opts.NumLevelZeroTablesStall = 10
		opts.NumMemtables = 5
		opts.SyncWrites = true
		opts.NumVersionsToKeep = 1
		opts.CompactL0OnClose = true
		opts.ValueLogFileSize = 1<<30 - 1
		opts.ValueLogMaxEntries = 1000000
		opts.ValueThreshold = 32
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("could not open new DB: %w", err)
	}

	return db, nil
}
