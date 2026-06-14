package query

import (
	"net/url"
	"testing"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name         string
		values       url.Values
		defaultLimit int
		maxLimit     int
		wantPage     int
		wantLimit    int
	}{
		{
			name:         "uses defaults",
			values:       url.Values{},
			defaultLimit: 20,
			maxLimit:     100,
			wantPage:     1,
			wantLimit:    20,
		},
		{
			name:         "parses valid values",
			values:       url.Values{"page": {"3"}, "limit": {"40"}},
			defaultLimit: 20,
			maxLimit:     100,
			wantPage:     3,
			wantLimit:    40,
		},
		{
			name:         "rejects non-positive values",
			values:       url.Values{"page": {"0"}, "limit": {"-1"}},
			defaultLimit: 20,
			maxLimit:     100,
			wantPage:     1,
			wantLimit:    20,
		},
		{
			name:         "caps the limit",
			values:       url.Values{"page": {"2"}, "limit": {"500"}},
			defaultLimit: 20,
			maxLimit:     100,
			wantPage:     2,
			wantLimit:    100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, limit := Pagination(tt.values, tt.defaultLimit, tt.maxLimit)
			if page != tt.wantPage || limit != tt.wantLimit {
				t.Fatalf("Pagination() = (%d, %d), want (%d, %d)", page, limit, tt.wantPage, tt.wantLimit)
			}
		})
	}
}
