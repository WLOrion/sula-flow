package usecase

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/WLOrion/sula-flow/internal/adapters"
	"github.com/WLOrion/sula-flow/internal/country"
	"github.com/WLOrion/sula-flow/internal/domain"
)

type TransferUsecase struct {
	scraper      adapters.IScraper
	jsonRepo     adapters.JSONRepository
	countryStore *country.CountryStore
}

func NewTransferUsecase(scraper adapters.IScraper, jsonRepo adapters.JSONRepository, cs *country.CountryStore) *TransferUsecase {
	return &TransferUsecase{
		scraper:      scraper,
		jsonRepo:     jsonRepo,
		countryStore: cs,
	}
}

func (uc *TransferUsecase) GetTransfers(countryID, fromYear, toYear int) ([]domain.Player, error) {
	country, ok := uc.countryStore.GetByID(countryID)
	if !ok {
		return nil, fmt.Errorf("invalid country id %d", countryID)
	}

	countryPath := sanitizeName(country)
	allPlayers := []domain.Player{}

	for year := fromYear; year <= toYear; year++ {
		fmt.Printf("============== %d ==============\n", year)

		players, err := uc.scraper.Scrape(country, countryID, year)
		if err != nil {
			return nil, err
		}

		allPlayers = append(allPlayers, players...)

		path := filepath.Join("transfers", countryPath, fmt.Sprintf("transfer_%s_%d.json", countryPath, year))
		if err := uc.jsonRepo.Save(path, players); err != nil {
			return nil, err
		}
		fmt.Println("")
	}

	return allPlayers, nil
}

// sanitizeName: transforma "Brasil/USA" em "Brasil_USA", remove caracteres inválidos
func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	return name
}
