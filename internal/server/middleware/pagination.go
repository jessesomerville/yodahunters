package middleware

import (
	"net/http"
	"strconv"
)

// A Page holds the metadata for pagination
type Page struct {
	Size   int
	Number int
}

func GetPageData(r *http.Request) (Page, error) {
	page := Page{
		Size:   20,
		Number: 1,
	}
	size_param := r.URL.Query().Get("page_size")
	if size_param != "" {
		size, err := strconv.Atoi(size_param)
		if err != nil {
			return page, err
		} else if size > 0 {
			page.Size = size
		}
	}

	number_param := r.URL.Query().Get("page_number")
	if number_param != "" {
		number, err := strconv.Atoi(number_param)
		if err != nil {
			return page, err
		} else if number > 0 {
			page.Number = number
		}
	}
	// we should return the default page values if the page_size or page_number aren't set
	return page, nil
}
