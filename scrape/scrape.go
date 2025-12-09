// Package scrape defines the Scraper interface for extracting job offer data from external sources.
// Implementations accept a query and return structured offer parameters for database insertion.
// Includes a mock implementation for testing.
package scrape

import (
	"errors"
	"net/http"

	"github.com/alwedo/jobber/db"
)

type Scraper interface {
	Scrape(*db.Query) ([]db.CreateOfferParams, error)
}

var ErrRetryable = errors.New("scrape: retryable error")

var isRetryable = map[int]bool{
	http.StatusRequestTimeout:      true,
	http.StatusTooEarly:            true,
	http.StatusTooManyRequests:     true,
	http.StatusInternalServerError: true,
	http.StatusBadGateway:          true,
	http.StatusServiceUnavailable:  true,
	http.StatusGatewayTimeout:      true,
}

type mockScraper struct {
	LastQuery *db.Query
}

func (m *mockScraper) Scrape(q *db.Query) ([]db.CreateOfferParams, error) {
	o := []db.CreateOfferParams{}
	m.LastQuery = q
	if q.Keywords == "retry" {
		return o, ErrRetryable
	}
	return o, nil
}

var MockScraper = &mockScraper{}
