package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/WLOrion/sula-flow/internal/core/usecase"
)

type HTTPHandler struct {
	ScrapeUC *usecase.ScrapeUC
}

func NewRouter(scrapeUC *usecase.ScrapeUC) *HTTPHandler {
	return &HTTPHandler{ScrapeUC: scrapeUC}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := usecase.ScrapeParams{
		Region:   r.URL.Query().Get("region"),
		FromYear: parseInt(r.URL.Query().Get("from"), 2023),
		ToYear:   parseInt(r.URL.Query().Get("to"), 2023),
	}

	transfers, err := h.ScrapeUC.Execute(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transfers)
}

func parseInt(s string, fallback int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return fallback
}
