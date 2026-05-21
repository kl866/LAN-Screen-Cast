package main

import (
	"flag"
	"log"

	"lan-screen-cast/internal/ui"
)

func main() {
	port := flag.String("port", ":9527", "listen address")
	flag.Parse()

	log.Printf("Starting viewer on %s", *port)
	if err := ui.RunViewer(*port); err != nil {
		log.Fatalf("viewer error: %v", err)
	}
}
