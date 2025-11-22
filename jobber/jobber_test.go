package jobber

import (
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/Alvaroalonsobabbel/jobber/db"
)

type mockScraper struct{}

func (m *mockScraper) scrape(*db.Query) ([]db.CreateOfferParams, error) {
	return []db.CreateOfferParams{}, nil
}

func TestCreateQuery(t *testing.T) {
	l := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	d, dbCloser := db.NewTestDB(t)
	defer dbCloser()
	j, jCloser := newConfigurableJobber(l, d, &mockScraper{})
	defer jCloser()

	t.Run("creates a query", func(t *testing.T) {
		q, err := j.CreateQuery("cuak", "squeek")
		if err != nil {
			t.Fatalf("failed to create query: %s", err)
		}
		if q.Keywords != "cuak" {
			t.Errorf("expected keywords to be 'cuak', got %s", q.Keywords)
		}
		if q.Location != "squeek" {
			t.Errorf("expected location to be 'squeek', got %s", q.Location)
		}
		if len(j.sched.Jobs()) != 4 { // 3 from the seed + the recently created.
			t.Errorf("expected number of jobs to be 4, got %d", len(j.sched.Jobs()))
		}
		time.Sleep(50 * time.Millisecond)
		for _, jb := range j.sched.Jobs() {
			if slices.Contains(jb.Tags(), q.Keywords+q.Location) {
				lr, _ := jb.LastRun() //nolint: errcheck
				if lr.Before(time.Now().Add(-time.Second)) {
					t.Errorf("expected created query to have been performed immediately, got %v", lr)
				}
			}
		}
	})

	t.Run("on existing query it returns the existing one", func(t *testing.T) {
		q, err := j.CreateQuery("golang", "Berlin")
		if err != nil {
			t.Fatalf("failed to create existing query: %s", err)
		}
		if q.ID != 5 {
			t.Errorf("expected query ID to be 5, got %d", q.ID)
		}
		if q.Keywords != "golang" {
			t.Errorf("expected keywords to be 'golang', got %s", q.Keywords)
		}
		if q.Location != "Berlin" {
			t.Errorf("expected location to be 'Berlin', got %s", q.Location)
		}
	})
}

func TestListOffers(t *testing.T) {
	l := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	d, dbCloser := db.NewTestDB(t)
	defer dbCloser()
	j, jCloser := newConfigurableJobber(l, d, &mockScraper{})
	defer jCloser()

	tests := []struct {
		name       string
		keywords   string
		location   string
		wantOffers int
		wantErr    error
	}{
		{
			name:       "valid query with offers",
			keywords:   "golang",
			location:   "berlin",
			wantOffers: 1,
			wantErr:    nil,
		},
		{
			name:       "valid query with older than 7 days offers",
			keywords:   "python",
			location:   "san francisco",
			wantOffers: 1, // query has two offers, one is older than 7 days.
		},
		{
			name:     "invalid query with no offers",
			keywords: "cuak",
			location: "squeek",
			wantErr:  sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := j.ListOffers(tt.keywords, tt.location)
			switch {
			case err == nil:
				if len(o) != tt.wantOffers {
					t.Errorf("expected %d offers, got %d", tt.wantOffers, len(o))
				}
			case errors.Is(err, tt.wantErr):
				// expected error
			default:
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
