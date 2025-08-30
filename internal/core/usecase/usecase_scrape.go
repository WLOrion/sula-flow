package usecase

import (
	"github.com/WLOrion/sula-flow/internal/adapters"
	"github.com/WLOrion/sula-flow/internal/domain"
)

type ScrapeUC struct {
	Scraper adapters.IScraper
}

func NewScrapeUC(scraper adapters.IScraper) *ScrapeUC {
	return &ScrapeUC{Scraper: scraper}
}

type ScrapeParams struct {
	Region   string
	FromYear int
	ToYear   int
}

func (uc *ScrapeUC) Execute(params ScrapeParams) ([]domain.Player, error) {
	return uc.Scraper.Scrape(params.Region, params.FromYear, params.ToYear)
}
