package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPageData(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantPage Page
		wantErr  bool
	}{
		{
			name:     "no query params returns defaults",
			query:    "",
			wantPage: Page{Size: 20, Number: 1},
		},
		{
			name:     "valid page_size and page_number",
			query:    "page_size=10&page_number=3",
			wantPage: Page{Size: 10, Number: 3},
		},
		{
			name:     "page_size greater than 100 capped to 100",
			query:    "page_size=200",
			wantPage: Page{Size: 100, Number: 1},
		},
		{
			name:     "page_size zero keeps default",
			query:    "page_size=0",
			wantPage: Page{Size: 20, Number: 1},
		},
		{
			name:     "page_size negative keeps default",
			query:    "page_size=-5",
			wantPage: Page{Size: 20, Number: 1},
		},
		{
			name:    "page_size non-numeric returns error",
			query:   "page_size=abc",
			wantErr: true,
		},
		{
			name:    "page_number non-numeric returns error",
			query:   "page_number=xyz",
			wantErr: true,
		},
		{
			name:     "page_number zero keeps default 1",
			query:    "page_number=0",
			wantPage: Page{Size: 20, Number: 1},
		},
		{
			name:     "page_number negative keeps default 1",
			query:    "page_number=-2",
			wantPage: Page{Size: 20, Number: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/test"
			if tt.query != "" {
				url += "?" + tt.query
			}
			r := httptest.NewRequest(http.MethodGet, url, nil)

			got, err := GetPageData(r)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Size != tt.wantPage.Size {
				t.Errorf("Size = %d, want %d", got.Size, tt.wantPage.Size)
			}
			if got.Number != tt.wantPage.Number {
				t.Errorf("Number = %d, want %d", got.Number, tt.wantPage.Number)
			}
		})
	}
}
