// The backend command runs the backend web server.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/google/safehtml/template"
	"github.com/jessesomerville/yodahunters/internal/server"
)

var (
	addr   = flag.String("addr", "localhost:8080", "the address for the server to listen on")
	static = flag.String("static", "static", "path to a directory containing static assets for the site")
)

func main() {
	flag.Parse()

	staticSrc := template.TrustedSourceFromFlag(flag.Lookup("static").Value)
	cfg := server.Config{
		Address:    *addr,
		TemplateFS: template.TrustedFSFromTrustedSource(staticSrc),
	}
	log.Fatal(server.Run(context.Background(), cfg))
}
