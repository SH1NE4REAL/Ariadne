package model

type PriceQuote struct {
	Type     string `json:"type"`
	Platform string `json:"platform"`
	Method   string `json:"method"`
	Price    int    `json:"price"`
	Duration string `json:"duration"`
	URL      string `json:"url"`
	Score    int    `json:"score"`
	Reason   string `json:"reason"`
}

type BestBookingOption struct {
	Best         PriceQuote   `json:"best"`
	Alternatives []PriceQuote `json:"alternatives"`
}