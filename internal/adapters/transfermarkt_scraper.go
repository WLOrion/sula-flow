package adapters

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/WLOrion/sula-flow/internal/domain"
	"github.com/WLOrion/sula-flow/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

type IScraper interface {
	Scrape(region string, countryID, year int) ([]domain.Player, error)
}

type TransfermarktScraper struct{}

func NewTransfermarktScraper() IScraper {
	return &TransfermarktScraper{}
}

func (s *TransfermarktScraper) Scrape(region string, countryID, year int) ([]domain.Player, error) {
	players := []domain.Player{}
	var lastFirstPlayerID int64

	for page := 1; page <= 75; page++ {
		fmt.Println("Scraping page:", page)
		url := fmt.Sprintf("https://www.transfermarkt.com/transfers/saisontransfers/statistik/top/plus/1/galerie/0?saison_id=%d&transferfenster=alle&land_id=%d&ausrichtung=&spielerposition_id=&altersklasse=&leihe=&page=%d", year, countryID, page)
		doc, err := fetchDocument(url)
		if err != nil {
			return nil, err
		}

		stopPage := false
		iter := doc.Find("table.items tbody tr").EachIter()

		numOfPlayersBefore := 0

		iter(func(idx int, row *goquery.Selection) bool {
			cols := row.Find("td")
			if cols.Length() < 17 {
				return true
			}

			// Checar jogador repetido na primeira linha
			if idx == 0 {
				firstPlayerOnPage, _ := strconv.ParseInt(cols.Eq(0).Text(), 10, 64)
				if firstPlayerOnPage == lastFirstPlayerID {
					stopPage = true
					return false
				}
				lastFirstPlayerID = firstPlayerOnPage
			}

			// PLAYER
			playerID := 0
			playerName := ""
			playerUrl := ""
			cols.Eq(1).Find("table.inline-table tbody tr").Each(func(i int, tr *goquery.Selection) {
				if i == 0 {
					link := tr.Find("td.hauptlink a")
					if link.Length() > 0 {
						href, exists := link.Attr("href")
						if exists {
							playerUrl = fmt.Sprintf("https://www.transfermarkt.com%s", href)
							parts := strings.Split(href, "/")
							if len(parts) > 0 {
								idStr := parts[len(parts)-1]
								playerID, _ = strconv.Atoi(idStr)
							}
							playerName = strings.TrimSpace(link.Text())
						}
					}
				}
			})

			if playerID == 0 || playerName == "" {
				return true
			}

			// FROM CLUB
			fromClub := domain.Club{}

			cols.Eq(10).Find("td.hauptlink a").Each(func(_ int, a *goquery.Selection) {
				href, exists := a.Attr("href")
				if exists {
					parts := strings.Split(href, "/")
					for i, p := range parts {
						if p == "verein" && i+1 < len(parts) {
							id, _ := strconv.Atoi(parts[i+1])
							fromClub.ClubID = id
							break
						}
					}
				}
				fromClub.ClubName = strings.TrimSpace(a.Text())
			})

			cols.Eq(11).Find("td img.flaggenrahmen").Each(func(_ int, a *goquery.Selection) {
				title, exists := a.Attr("title")
				if exists {
					fromClub.Country = title
				}
			})

			// TO CLUB
			toClub := domain.Club{}
			cols.Eq(14).Find("td.hauptlink a").Each(func(_ int, a *goquery.Selection) {
				href, exists := a.Attr("href")
				if exists {
					parts := strings.Split(href, "/")
					for i, p := range parts {
						if p == "verein" && i+1 < len(parts) {
							id, _ := strconv.Atoi(parts[i+1])
							toClub.ClubID = id
							break
						}
					}
				}
				toClub.ClubName = strings.TrimSpace(a.Text())
			})

			cols.Eq(15).Find("td img.flaggenrahmen").Each(func(_ int, a *goquery.Selection) {
				title, exists := a.Attr("title")
				if exists {
					toClub.Country = title
				}
			})

			// FEE
			feeText := strings.TrimSpace(cols.Eq(16).Text())
			fee, isLoan := utils.ParseFee(feeText)

			players = append(players, domain.Player{
				PlayerID:    playerID,
				PlayerName:  playerName,
				PlayerUrl:   playerUrl,
				Nationality: region,
				Transfer: domain.Transfer{
					From:   fromClub,
					To:     toClub,
					FeeEUR: fee,
					IsLoan: isLoan,
					Season: fmt.Sprintf("%d/%d", year, year+1),
				},
			})

			return true
		})

		if stopPage || numOfPlayersBefore == len(players) {
			break
		}
	}

	return players, nil
}

func fetchDocument(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
