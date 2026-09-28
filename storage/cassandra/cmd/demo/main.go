// Command demo writes one timeline row and reads it back at QUORUM:
// same shape as schema.cql, driver-enforced consistency per query.
// Usage: go run ./cmd/demo (needs make up + schema applied).
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gocql/gocql"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "demo:", err)

	os.Exit(1)
}

func main() {
	cluster := gocql.NewCluster(environ.Get("CASS_HOST", "localhost"))
	cluster.Keyspace = "lab"
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		fail(err)
	}

	defer session.Close()

	userID := 7

	err = session.Query(
		`INSERT INTO timeline (user_id, created_at, post_id, body) VALUES (?, ?, uuid(), ?)`,
		userID, time.Now().UTC(), "hello cassandra",
	).Exec()
	if err != nil {
		fail(err)
	}

	fmt.Println("wrote 1 row at QUORUM")

	iter := session.Query(
		`SELECT created_at, body FROM timeline WHERE user_id = ? LIMIT 5`,
		userID,
	).Consistency(gocql.One).Iter()

	var at time.Time

	var body string

	count := 0

	for iter.Scan(&at, &body) {
		fmt.Printf("row: %s %q\n", at.Format(time.RFC3339), body)

		count++
	}

	err = iter.Close()
	if err != nil {
		fail(err)
	}

	fmt.Printf("read %d rows at ONE (stale-ok)\n", count)
}
