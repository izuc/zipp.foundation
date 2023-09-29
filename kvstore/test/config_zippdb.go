//go:build zippdb
// +build zippdb

package test

var (
	dbImplementations = []string{"badger", "mapDB", "pebble", "zippdb"}
)
