package models

// SensexIndex holds the top-level Sensex index data from BSE API
type SensexIndex struct {
	Index        string  `json:"index"`
	Last         float64 `json:"last"`
	Change       float64 `json:"change"`
	PercentChange float64 `json:"percentChange"`
	Open         float64 `json:"open"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
	PreviousClose float64 `json:"previousClose"`
	Timestamp    string  `json:"timestamp"`
}

// BSEIndexAPIResponse is what BSE equity-stockIndices returns
type BSEIndexAPIResponse struct {
	Data []BSEIndexData `json:"data"`
}

type BSEIndexData struct {
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
}

// Constituent holds a single Sensex constituent stock with its weightage
type Constituent struct {
	Rank         int     `json:"rank"`
	ScripCode    string  `json:"scripCode"`
	CompanyName  string  `json:"companyName"`
	Weightage    float64 `json:"weightage"`   // % weightage in Sensex
	PointsChange float64 `json:"pointsChange"` // Calculated: sensexChange * (weightage/100)
	Last         float64 `json:"last"`
	Change       float64 `json:"change"`
	PercentChange float64 `json:"percentChange"`
}

// BSEWeightageResponse is parsed from BSE Sensex constituents page
type BSEWeightageItem struct {
	ScripCode   string  `json:"scrip_cd"`
	CompanyName string  `json:"company_name"`
	Weightage   float64 `json:"weightage"`
}

// SensexSnapshot is the full response sent to the frontend
type SensexSnapshot struct {
	Index        SensexIndex   `json:"index"`
	Constituents []Constituent `json:"constituents"`
	LastUpdated  string        `json:"lastUpdated"`
	Error        string        `json:"error,omitempty"`
}
