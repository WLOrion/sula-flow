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
	scraper   adapters.IScraper
	jsonRepo  adapters.JSONRepository
	countries *country.Country
}

func NewTransferUsecase(scraper adapters.IScraper, jsonRepo adapters.JSONRepository, countries *country.Country) *TransferUsecase {
	return &TransferUsecase{
		scraper:   scraper,
		jsonRepo:  jsonRepo,
		countries: countries,
	}
}

func (uc *TransferUsecase) GetTransfers(countryID, fromYear, toYear int) ([]domain.Player, error) {
	region, err := uc.countries.NameByID(countryID)
	if err != nil {
		return nil, err
	}

	// Formata o nome do país para path seguro
	regionPath := sanitizeName(region)

	allPlayers := []domain.Player{}

	for year := fromYear; year <= toYear; year++ {
		fmt.Printf("============== %d ==============\n", year)

		players, err := uc.scraper.Scrape(region, countryID, year)
		if err != nil {
			return nil, err
		}

		allPlayers = append(allPlayers, players...)

		// Salva JSON por ano
		path := filepath.Join("transfers", regionPath, fmt.Sprintf("transfer_%s_%d.json", regionPath, year))
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
