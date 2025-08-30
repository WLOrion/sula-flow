package main

import (
	"log"
	"net/http"

	"github.com/WLOrion/sula-flow/internal/adapters"
	"github.com/WLOrion/sula-flow/internal/core/usecase"
	httpp "github.com/WLOrion/sula-flow/internal/ports/http"
)

func main() {
	scraper := adapters.NewTransfermarktScraper()
	scrapeUC := usecase.NewScrapeUC(scraper)
	handler := httpp.NewRouter(scrapeUC)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Server running at http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
