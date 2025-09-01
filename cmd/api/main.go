package main

import (
	"log"
	"net/http"

	"github.com/WLOrion/sula-flow/internal/country"
	router "github.com/WLOrion/sula-flow/internal/ports/http"
)

func main() {
	countryStore, err := country.LoadCountries("docs/csv/countries.csv")
	if err != nil {
		log.Fatalf("failed to load countries: %v", err)
	}

	handler := router.NewRouter(countryStore)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Server running at http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
