package currencybeacon

type fetchCurrenciesResponse struct {
	Response []currency `json:"response"`
}

type currency struct {
	Name      string `json:"name"`
	ShortCode string `json:"short_code"`
	Symbol    string `json:"symbol"`
}

type getExchangeRateResponse struct {
	Rates map[string]float64 `json:"rates"`
}
