package main

import (
	"log"
	"net/http"

	"github.com/WLOrion/sula-flow/internal/country"
	router "github.com/WLOrion/sula-flow/internal/ports/http"
)

func main() {
	// Carrega CSV de países
	countries, err := country.LoadCountries("docs/csv/countries.csv")
	if err != nil {
		log.Fatal(err)
	}

	handler := router.NewRouter(countries)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Server running at http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
