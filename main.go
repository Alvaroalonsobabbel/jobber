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
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "modernc.org/sqlite"
)

const (
	logFile = "jobber.log"
	dbFile  = "jobber.sqlite"

	stateQueries = iota
	stateNewQuery
	stateOffers
	stateViewOffer
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

var queryColumns = []table.Column{
	{Title: "Queries", Width: 20},
}

var offerColumns = []table.Column{
	{Title: "Title", Width: 40},
	{Title: "Location", Width: 10},
	{Title: "Company", Width: 30},
	{Title: "Posted", Width: 10},
}

func newTable(c []table.Column) table.Model {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t := table.New(
		table.WithColumns(c),
		table.WithFocused(false),
		table.WithHeight(7),
	)
	t.SetStyles(s)

	return t
}

type model struct {
	jobber       *jobber.Jobber
	queriesTable table.Model
	offersTable  table.Model
	queries      []*db.Query
	offers       map[int64][]*db.Offer
	state        int
}

func NewModel(j *jobber.Jobber) *model {
	return &model{
		jobber:       j,
		queriesTable: newTable(queryColumns),
		offersTable:  newTable(offerColumns),
		offers:       make(map[int64][]*db.Offer),
		state:        stateQueries,
	}
}

func (m *model) Init() tea.Cmd {
	m.buildQueries()
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.setState(stateQueries)
			return m, nil
		case "n":
			if m.state == stateQueries {

			}
		case "enter":
			m.setState(stateOffers)
			return m, m.performSearch(m.queries[m.queriesTable.Cursor()])
		}
	}
	m.queriesTable.Blur()
	m.offersTable.Blur()

	var cmd tea.Cmd
	switch m.state {
	case stateQueries:
		m.queriesTable.Focus()
		m.queriesTable, cmd = m.queriesTable.Update(msg)
	case stateOffers:
		m.offersTable.Focus()
		m.offersTable, cmd = m.offersTable.Update(msg)
	}
	return m, cmd
}

func (m *model) View() string {
	var body string
	switch m.state {
	case stateQueries:
		body = m.queriesTable.View()
	case stateOffers:
		body = m.offersTable.View()
	}
	return baseStyle.Render(body + "\n" + "show help here")
}

func (m *model) setState(s int) {
	m.state = s
}

func (m *model) performSearch(q *db.Query) tea.Cmd {
	m.offers[q.ID] = m.jobber.RunQuery(q)
	var rows []table.Row
	for _, o := range m.offers[q.ID] {
		row := table.Row{o.Title, o.Location, o.Company, o.PostedAt.Format("01/02")}
		rows = append(rows, row)
	}
	m.offersTable.SetRows(rows)

	// TODO: return command for spinner
	return nil
}

func (m *model) buildQueries() {
	m.queries = m.jobber.ListQueries()
	var rows []table.Row
	for _, q := range m.queries {
		row := table.Row{q.Keywords + " " + q.Location}
		rows = append(rows, row)
	}
	m.queriesTable.SetRows(rows)
}

func main() {
	f, err := tea.LogToFile(logFile, "teaDebug")
	if err != nil {
		log.Fatal("fatal:", err)
	}
	defer f.Close()

	d, closeDB := initDB()
	defer closeDB.Close()

	logger, closeLogger := initLogger()
	defer closeLogger.Close()

	j := jobber.New(logger, d)
	m := NewModel(j)
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

//go:embed schema.sql
var ddl string

func initLogger() (*slog.Logger, io.Closer) {
	// TODO: change file location to home folder
	out, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("unable to open log file: %v", err)
	}

	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(handler), out
}

func initDB() (*db.Queries, io.Closer) {
	// TODO: change file location to home folder
	d, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("unable to open database file: %v", err)
	}
	if _, err := d.ExecContext(context.Background(), ddl); err != nil {
		log.Fatalf("unable to create database schema: %v", err)
	}
	return db.New(d), d
}
