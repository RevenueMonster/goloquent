package examples

import (
	"time"

	"cloud.google.com/go/datastore"
	"github.com/RevenueMonster/goloquent"
)

// User :
type User struct {
	Key       *datastore.Key `goloquent:"__key__"`
	Name      string
	Status    string
	CreatedAt time.Time
	Deleted   goloquent.SoftDelete
}
