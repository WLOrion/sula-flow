package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/WLOrion/sula-flow/internal/adapters"
	"github.com/WLOrion/sula-flow/internal/core/usecase"
	"github.com/WLOrion/sula-flow/internal/country"
)

func NewRouter(cs *country.CountryStore) http.Handler {
	mux := http.NewServeMux()

	scraper := adapters.NewTransfermarktScraper()
	repo := adapters.NewJSONRepository()

	transferUC := usecase.NewTransferUsecase(scraper, repo, cs)
	unifyUC := usecase.NewUnifyUsecase(cs)

	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		countryIDStr := r.URL.Query().Get("country_id")
		countryID, err := strconv.Atoi(countryIDStr)
		if err != nil {
			http.Error(w, "invalid country id", http.StatusBadRequest)
			return
		}

		fromYearStr := r.URL.Query().Get("from")
		fromYear, _ := strconv.Atoi(fromYearStr)
		if err != nil {
			http.Error(w, "invalid from year", http.StatusBadRequest)
			return
		}

		toYearStr := r.URL.Query().Get("to")
		toYear, _ := strconv.Atoi(toYearStr)
		if err != nil {
			http.Error(w, "invalid to year", http.StatusBadRequest)
			return
		}

		_, ok := cs.GetByID(countryID)
		if !ok {
			http.Error(w, "country not found", http.StatusNotFound)
			return
		}

		transfers, err := transferUC.GetTransfers(countryID, fromYear, toYear)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transfers)
	})

	// POST /unifies/<COUNTRY-ID>
	mux.HandleFunc("/unifies/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// extrai o countryID da URL
		pathParts := r.URL.Path[len("/unifies/"):]
		countryID, err := strconv.Atoi(pathParts)
		if err != nil {
			http.Error(w, "invalid country id", http.StatusBadRequest)
			return
		}

		_, ok := cs.GetByID(countryID)
		if !ok {
			http.Error(w, "country not found", http.StatusNotFound)
			return
		}

		totalTransfers, uniquePlayers, err := unifyUC.UnifyTransfers(countryID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := map[string]int{
			"total_transfers": totalTransfers,
			"unique_players":  uniquePlayers,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	return mux
}
