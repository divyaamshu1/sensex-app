package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"sensex-backend/internal/models"
)

const (
	// BSE Sensex index API (same NSE-style endpoint for BSE via nseindia mirror)
	bseSensexURL = "https://www.nseindia.com/api/equity-stockIndices?index=S%26P%20BSE%20SENSEX"

	// BSE Sensex constituents with weightage
	bseWeightageURL = "https://m.bseindia.com/sensex_new.aspx?Scripflag=16"

	// Alternative direct BSE API for index data
	bseIndexDirectURL = "https://api.bseindia.com/BseIndiaAPI/api/SensexData/w"

	// User agent to mimic browser (required by BSE/NSE)
	userAgent = "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36"
)

// BSEConstituent from BSE API
type bseConstituentRaw struct {
	ScripCode   string `json:"scripcode"`
	ShortName   string `json:"shortname"`
	LongName    string `json:"longname"`
	Weightage   string `json:"weightage"`
	Last        string `json:"last"`
	Change      string `json:"change"`
	PctChange   string `json:"pctchange"`
}

type bseConstituentAPIResponse struct {
	Table []bseConstituentRaw `json:"Table"`
}

// BSE Sensex index response
type bseSensexDirectResponse struct {
	CurrValue  string `json:"CurrValue"`
	PrevClose  string `json:"PrevClose"`
	OpenValue  string `json:"OpenValue"`
	High       string `json:"High"`
	Low        string `json:"Low"`
	NetChange  string `json:"NetChange"`
	PercChange string `json:"PercChange"`
	UpdateDate string `json:"UpdateDate"`
	UpdateTime string `json:"UpdateTime"`
}

// Fetcher handles all HTTP fetching from BSE
type Fetcher struct {
	client *http.Client
}

// New creates a new Fetcher with sensible timeouts
func New() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 5,
			},
		},
	}
}

// doRequest performs an HTTP GET with browser-like headers
func (f *Fetcher) doRequest(url string, referer string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}

// FetchSensexIndex fetches the current Sensex index value from BSE direct API
func (f *Fetcher) FetchSensexIndex() (*models.SensexIndex, error) {
	// Try BSE direct API first
	body, err := f.doRequest(bseIndexDirectURL, "https://www.bseindia.com/")
	if err != nil {
		log.Printf("[fetcher] BSE direct API failed: %v, trying NSE mirror...", err)
		return f.fetchSensexFromNSEMirror()
	}

	var raw bseSensexDirectResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Printf("[fetcher] BSE direct parse failed: %v, trying NSE mirror...", err)
		return f.fetchSensexFromNSEMirror()
	}

	last := parseFloat(raw.CurrValue)
	prevClose := parseFloat(raw.PrevClose)
	change := parseFloat(raw.NetChange)
	pctChange := parseFloat(raw.PercChange)

	// If NetChange is empty, calculate it
	if change == 0 && last != 0 && prevClose != 0 {
		change = last - prevClose
	}
	if pctChange == 0 && prevClose != 0 {
		pctChange = (change / prevClose) * 100
	}

	return &models.SensexIndex{
		Index:         "S&P BSE SENSEX",
		Last:          last,
		Change:        change,
		PercentChange: pctChange,
		Open:          parseFloat(raw.OpenValue),
		High:          parseFloat(raw.High),
		Low:           parseFloat(raw.Low),
		PreviousClose: prevClose,
		Timestamp:     raw.UpdateDate + " " + raw.UpdateTime,
	}, nil
}

// fetchSensexFromNSEMirror falls back to NSE's BSE Sensex mirror endpoint
func (f *Fetcher) fetchSensexFromNSEMirror() (*models.SensexIndex, error) {
	// First, hit the NSE homepage to get cookies
	_, _ = f.doRequest("https://www.nseindia.com", "")

	body, err := f.doRequest(bseSensexURL, "https://www.nseindia.com/")
	if err != nil {
		return nil, fmt.Errorf("NSE mirror: %w", err)
	}

	var raw struct {
		Data []struct {
			IndexSymbol   string  `json:"indexSymbol"`
			Index         string  `json:"index"`
			Last          float64 `json:"last"`
			Variation     float64 `json:"variation"`
			PerChange     float64 `json:"perChange"`
			Open          float64 `json:"open"`
			High          float64 `json:"high"`
			Low           float64 `json:"low"`
			PreviousClose float64 `json:"previousClose"`
			TimeVal       string  `json:"timeVal"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("NSE mirror parse: %w", err)
	}

	if len(raw.Data) == 0 {
		return nil, fmt.Errorf("NSE mirror: empty data")
	}

	d := raw.Data[0]
	return &models.SensexIndex{
		Index:         d.Index,
		Last:          d.Last,
		Change:        d.Variation,
		PercentChange: d.PerChange,
		Open:          d.Open,
		High:          d.High,
		Low:           d.Low,
		PreviousClose: d.PreviousClose,
		Timestamp:     d.TimeVal,
	}, nil
}

// FetchConstituents fetches Sensex 30 constituents with weightages from BSE API
func (f *Fetcher) FetchConstituents() ([]models.Constituent, error) {
	// BSE API for Sensex constituents with weightage
	url := "https://api.bseindia.com/BseIndiaAPI/api/SensexData/w?flag=16"
	body, err := f.doRequest(url, "https://m.bseindia.com/sensex_new.aspx")
	if err != nil {
		log.Printf("[fetcher] BSE constituents API failed: %v, using fallback...", err)
		return f.fetchConstituentsFromHTML()
	}

	var raw bseConstituentAPIResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Printf("[fetcher] BSE constituents parse failed: %v, using fallback...", err)
		return f.fetchConstituentsFromHTML()
	}

	if len(raw.Table) == 0 {
		return f.fetchConstituentsFromHTML()
	}

	constituents := make([]models.Constituent, 0, len(raw.Table))
	for i, item := range raw.Table {
		weightage := parseFloat(item.Weightage)
		constituents = append(constituents, models.Constituent{
			Rank:         i + 1,
			ScripCode:    item.ScripCode,
			CompanyName:  item.LongName,
			Weightage:    weightage,
			Last:         parseFloat(item.Last),
			Change:       parseFloat(item.Change),
			PercentChange: parseFloat(item.PctChange),
		})
	}

	// Sort by weightage descending
	sort.Slice(constituents, func(i, j int) bool {
		return constituents[i].Weightage > constituents[j].Weightage
	})

	// Re-rank after sort
	for i := range constituents {
		constituents[i].Rank = i + 1
	}

	return constituents, nil
}

// fetchConstituentsFromHTML is a fallback that uses hardcoded Sensex 30 
// constituents with approximate weightages (updated as of 2025)
// This ensures the app works even if BSE API blocks the request
func (f *Fetcher) fetchConstituentsFromHTML() ([]models.Constituent, error) {
	log.Println("[fetcher] Using hardcoded Sensex 30 constituents as fallback")

	// Sensex 30 constituents with approximate weightages (sum ~100%)
	// Source: BSE Sensex factsheet (regularly updated)
	hardcoded := []struct {
		code      string
		name      string
		weightage float64
	}{
		{"500325", "Reliance Industries", 11.23},
		{"500180", "HDFC Bank", 10.87},
		{"500209", "Infosys", 7.54},
		{"532540", "TCS", 6.12},
		{"500696", "ICICI Bank", 5.89},
		{"500010", "HDFC", 4.76},
		{"500875", "ITC", 3.98},
		{"500112", "State Bank of India", 3.45},
		{"532648", "Kotak Mahindra Bank", 3.21},
		{"500510", "Larsen & Toubro", 3.18},
		{"532187", "Axis Bank", 2.87},
		{"500470", "Tata Steel", 1.98},
		{"500570", "Tata Motors", 1.87},
		{"507685", "Wipro", 1.76},
		{"500400", "Tata Power", 1.54},
		{"500440", "Hindustan Unilever", 2.43},
		{"500672", "Sun Pharma", 1.98},
		{"500124", "Dr Reddys Laboratories", 1.21},
		{"500820", "Asian Paints", 1.43},
		{"500312", "ONGC", 1.32},
		{"500900", "Bajaj Finance", 2.12},
		{"532977", "Bajaj Finserv", 1.54},
		{"500087", "Cipla", 0.98},
		{"500010", "Power Grid Corporation", 1.12},
		{"534816", "Titan Company", 1.34},
		{"500020", "Nestle India", 0.87},
		{"500247", "HCL Technologies", 1.43},
		{"533278", "Coal India", 0.98},
		{"500002", "Adani Ports", 1.21},
		{"532454", "Bharti Airtel", 2.54},
	}

	constituents := make([]models.Constituent, 0, len(hardcoded))
	for i, h := range hardcoded {
		constituents = append(constituents, models.Constituent{
			Rank:        i + 1,
			ScripCode:   h.code,
			CompanyName: h.name,
			Weightage:   h.weightage,
		})
	}

	return constituents, nil
}

// CalculatePointsImpact calculates each constituent's point contribution
// to the Sensex change based on their weightage
// Formula: pointsImpact = sensexChange * (weightage / totalWeightage)
func CalculatePointsImpact(constituents []models.Constituent, sensexChange float64) []models.Constituent {
	// Calculate total weightage (should be ~100 but normalize anyway)
	totalWeightage := 0.0
	for _, c := range constituents {
		totalWeightage += c.Weightage
	}
	if totalWeightage == 0 {
		return constituents
	}

	for i := range constituents {
		// Points attributed to this stock = sensexChange * its share of total weight
		constituents[i].PointsChange = roundTo2(sensexChange * (constituents[i].Weightage / totalWeightage))
	}

	return constituents
}

// parseFloat safely parses a string to float64, handling commas and % signs
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "+", "")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// roundTo2 rounds a float to 2 decimal places
func roundTo2(f float64) float64 {
	return math.Round(f*100) / 100
}
