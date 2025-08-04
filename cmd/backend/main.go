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
	addr    = flag.String("addr", "localhost:8080", "the address for the server to listen on")
	devmode = flag.Bool("devmode", false, "enable devmode (reload templates on each page load)")
)

func main() {
	flag.Parse()

	staticSrc := template.TrustedSourceFromConstant("static")
	cfg := server.Config{
		Address:    *addr,
		TemplateFS: template.TrustedFSFromTrustedSource(staticSrc),
		DevMode:    *devmode,
	}
	log.Fatal(server.Run(context.Background(), cfg))
}
