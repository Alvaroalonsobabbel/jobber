package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/Alvaroalonsobabbel/jobber/db"
	"github.com/Alvaroalonsobabbel/jobber/jobber"
	_ "modernc.org/sqlite"
)

func main() {
	logger, closer := initLogger()
	defer closer.Close()

	d, closer := initDB()
	defer closer.Close()

	jb := jobber.New(logger, d)

	query := &db.Query{}
	queries := jb.ListQueries()
	if len(queries) == 0 {
		query = jb.NewQuery(&db.CreateQueryParams{
			Keywords: "golang",
			Location: "berlin",
			FTpr:     jobber.ThreeDaysAgo,
		})
	} else {
		query = queries[0]
	}

	offers := jb.RunQuery(query)
	fmt.Printf("We got %d offers.\n\n", len(offers))
	for _, o := range offers {
		fmt.Printf("- %s at %s posted on %s\n", o.Title, o.Company, o.PostedAt.Format("2006-01-02"))
	}
}

//go:embed schema.sql
var ddl string

func initLogger() (*slog.Logger, io.Closer) {
	// TODO: change file location to home folder
	out, err := os.OpenFile("jobber.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("unable to open log file: %v", err)
	}

	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(handler), out
}

func initDB() (*db.Queries, io.Closer) {
	// TODO: change file location to home folder
	d, err := sql.Open("sqlite", "jobber.sqlite")
	if err != nil {
		log.Fatalf("unable to open database: %v", err)
	}
	if _, err := d.ExecContext(context.Background(), ddl); err != nil {
		log.Fatalf("unable to create database: %v", err)
	}
	return db.New(d), d
}
