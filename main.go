package main

import (
	"context"
	"database/sql"
	_ "embed"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Alvaroalonsobabbel/jobber/db"
	"github.com/Alvaroalonsobabbel/jobber/jobber"
	"github.com/Alvaroalonsobabbel/jobber/server"
	_ "modernc.org/sqlite"
)

func main() {
	logger, closer := initLogger()
	defer closer.Close()

	d, closer := initDB()
	defer closer.Close()

	j := jobber.New(logger, d)

	svr := server.New(logger, j)
	log.Println("starting server in port 80")
	if err := http.ListenAndServe(":80", svr); err != nil {
		log.Fatal(err)
	}
}

//go:embed schema.sql
var ddl string

func initLogger() (*slog.Logger, io.Closer) {
	out, err := os.OpenFile("jobber.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("unable to open log file: %v", err)
	}

	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(handler), out
}

func initDB() (*db.Queries, io.Closer) {
	d, err := sql.Open("sqlite", "jobber.sqlite")
	if err != nil {
		log.Fatalf("unable to open database: %v", err)
	}
	if _, err := d.ExecContext(context.Background(), ddl); err != nil {
		log.Fatalf("unable to create database: %v", err)
	}
	return db.New(d), d
}
