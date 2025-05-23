package examples

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/RevenueMonster/goloquent"
	"github.com/RevenueMonster/goloquent/db"
	"github.com/RevenueMonster/goloquent/expr"
	"github.com/brianvoe/gofakeit"

	//  "database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func TestExamples(t *testing.T) {

	ctx := context.Background()
	// mysql.RegisterTLSConfig("custom", &tls.Config{})
	conn, err := db.Open(ctx, "mysql", db.Config{
		Username: "root",
		Password: "abcd1234",
		Host:     "localhost",
		Port:     "3306",
		// TLSConfig: "",
		Database: "goloquent",
		Logger: func(ctx context.Context, stmt *goloquent.Stmt) {
			log.Println(stmt.TimeElapse()) // elapse time in time.Duration
			log.Println(stmt.String())     // Sql string without any ?
			log.Println(stmt.Raw())        // Sql prepare statement
			log.Println(stmt.Arguments())  // Sql prepare statement's arguments
			log.Println(fmt.Sprintf("[%.3fms] %s", stmt.TimeElapse().Seconds()*1000, stmt.String()))
		},
	})
	// defer conn.Close()
	if err != nil {
		panic(err)
	}

	db.Migrate(ctx, new(User))
	db.Truncate(ctx, "User")
	u := new(User)
	err = db.MatchAgainst([]string{"Name", "Username"}, "value", "value2").Find(ctx, datastore.IDKey("test", 100, nil), u)
	log.Println(err)

	users := [...]User{
		newUser(),
		newUser(),
		newUser(),
		newUser(),
		newUser(),
	}
	db.Create(ctx, &users)
	usrs := []User{}
	db.NewQuery().OrderBy(
		expr.Field("Status", []string{
			"A", "B", "C",
		}),
		"-CreatedAt",
	).Get(ctx, &usrs)

	query := db.NewQuery().OrderBy(
		"-CreatedAt",
	)
	pg := &goloquent.Pagination{Limit: 1}
	err = query.Paginate(ctx, pg, &usrs)

	pg.Cursor = pg.NextCursor()
	err = query.Paginate(ctx, pg, &usrs)

	pg.Cursor = pg.NextCursor()
	err = query.Paginate(ctx, pg, &usrs)

	log.Println(err)
	log.Println(usrs)
	// db.Create()
	log.Println(conn)
}

func newUser() (u User) {
	u.Name = ""
	u.CreatedAt = time.Now()
	u.Status = gofakeit.RandString([]string{
		"A",
		"B",
		"C",
	})
	return
}
