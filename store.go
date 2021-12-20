package goplumber

import (
	"github.com/philippgille/gokv"
	gokvmap "github.com/philippgille/gokv/gomap"
)

// Store is the type alias of gokv.Store.
type Store = gokv.Store

// DefaultStore returns the default store that is local in-memory store.
// This is the default. so, it should not be used for the real situation.
func DefaultStore() Store {
	return gokvmap.NewStore(gokvmap.DefaultOptions)
}
