// Package markdown implements a basic markdown syntax parser.
package markdown

import (
	md "rsc.io/markdown"
)

var parser = md.Parser{
	AutoLinkText: true,
	Table:        true,
}

func MarkdownToHTML(s string) string {
	doc := parser.Parse(s)
	return md.ToHTML(doc)
}
