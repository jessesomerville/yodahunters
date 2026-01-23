package templates

import (
	"html/template"
)

// Placing evil template functions in here
func renderQuoteBlock(commentBody string) template.HTML {
	// quotePattern := regexp.MustCompile(`\[QUOTE PAGE_NUMBER=(\d+) PAGE_SIZE=(\d+) COMMENT_ID=comment-(\d+)\](.+?)\[/QUOTE\]`)
	// replaceStr := `
	// <div class="quote-block">
	// 	<p class="quote-heading">in reply to this <a href="?page_number=$1&page_size=$2#comment-$3">comment</a></p>
	// 	<p class="quote-body">$4</p>
	// </div>`

	return template.HTML("<p>hello</p>")
}
