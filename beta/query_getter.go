package goloquent

import "cloud.google.com/go/datastore"

// Getter :
type Getter struct {
	query *Query
}

func newGetter(q *Query) *Getter {
	return &Getter{
		query: q,
	}
}

// Get :
func (g *Getter) Get(modelStruct interface{}) error {
	return newBuilder(g.query).Get(modelStruct)
}

// First :
func (g *Getter) First(modelStruct interface{}) error {
	return newBuilder(g.query).First(modelStruct)
}

// Find :
func (g *Getter) Find(key *datastore.Key, modelStruct interface{}) error {
	return newBuilder(g.query).Find(key, modelStruct)
}
