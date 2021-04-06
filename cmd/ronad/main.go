package main

import (
	"fmt"
	"os"

	"github.com/richardmarbach/rona/http"
	"github.com/richardmarbach/rona/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	db := sqlite.NewDB(":memory:")

	if err := db.Open(); err != nil {
		return err
	}
	quickTestService := sqlite.NewQuickTestService(db)
	server, err := http.NewServer(quickTestService)
	if err != nil {
		return err
	}

	return server.Start()
}
