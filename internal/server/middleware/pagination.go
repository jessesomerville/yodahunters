package middleware

import (
	"net/http"
	"strconv"
)

// A Page holds the metadata for pagination.
type Page struct {
	Size   int
	Number int
}

// GetPageData retrieves the page_size and page_number query parameters from
// the requested URL.
func GetPageData(r *http.Request) (Page, error) {
	page := Page{
		Size:   20,
		Number: 1,
	}
	sizeParam := r.URL.Query().Get("page_size")
	if sizeParam != "" {
		size, err := strconv.Atoi(sizeParam)
		if err != nil {
			return page, err
		} else if size > 0 && size < 100 {
			page.Size = size
		} else if size > 100 {
			page.Size = 100
		}
	}

	numberParam := r.URL.Query().Get("page_number")
	if numberParam != "" {
		number, err := strconv.Atoi(numberParam)
		if err != nil {
			return page, err
		} else if number > 0 {
			page.Number = number
		}
	}
	// we should return the default page values if the page_size or page_number aren't set
	return page, nil
}
