package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/WLOrion/sula-flow/internal/adapters"
	"github.com/WLOrion/sula-flow/internal/core/usecase"
	"github.com/WLOrion/sula-flow/internal/country"
)

func NewRouter(countries *country.Country) http.Handler {
	mux := http.NewServeMux()

	// Cria instâncias do scraper e do repositório
	scraper := adapters.NewTransfermarktScraper()
	repo := adapters.NewJSONRepository()
	transferUC := usecase.NewTransferUsecase(scraper, repo, countries)

	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		countryIDStr := r.URL.Query().Get("country_id")
		countryID, err := strconv.Atoi(countryIDStr)
		if err != nil {
			http.Error(w, "Invalid country_id", http.StatusBadRequest)
			return
		}

		// Valida se o country_id existe
		if _, err := countries.NameByID(countryID); err != nil {
			http.Error(w, "Country not found", http.StatusNotFound)
			return
		}

		fromYearStr := r.URL.Query().Get("from")
		toYearStr := r.URL.Query().Get("to")

		fromYear, _ := strconv.Atoi(fromYearStr)
		toYear, _ := strconv.Atoi(toYearStr)

		transfers, err := transferUC.GetTransfers(countryID, fromYear, toYear)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transfers)
	})

	return mux
}
