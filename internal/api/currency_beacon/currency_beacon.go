package currencybeacon

import (
	"fmt"
	"net/http"

	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"resty.dev/v3"
)

type currencyBeacon struct {
	httpClient *resty.Client
}

// New creates a new instance of currencyBeacon api.
func New(apiURL, apiKey string) *currencyBeacon {
	httpClient := resty.New().
		SetBaseURL(apiURL).
		SetAuthScheme("Bearer").
		SetAuthToken(apiKey)

	return &currencyBeacon{
		httpClient: httpClient,
	}
}

func (c *currencyBeacon) FetchCurrencies() ([]service.Currency, error) {
	var result fetchCurrenciesResponse

	response, err := c.httpClient.R().
		SetResult(&result).
		Get("/v1/currencies")
	if err != nil {
		return nil, fmt.Errorf("send fetch currencies request: %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("could not fetch currencies(statusCode: %d, body:%s)", response.StatusCode(), response.String())
	}

	output := make([]service.Currency, 0, len(result.Response))
	for _, currency := range result.Response {
		output = append(output, service.Currency{
			Name:   currency.Name,
			Code:   currency.ShortCode,
			Symbol: currency.Symbol,
		})
	}

	return output, nil
}

func (c *currencyBeacon) GetExchangeRate(baseCurrency, targetCurrency string) (*money.Money, error) {
	var result getExchangeRateResponse

	response, err := c.httpClient.R().
		SetResult(&result).
		Get(fmt.Sprintf("/v1/latest?base=%s", baseCurrency))
	if err != nil {
		return nil, fmt.Errorf("send get exchange rate request: %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("could not get exchange rate(statusCode: %d, body:%s)", response.StatusCode(), response.String())
	}

	exchangeRate, ok := result.Rates[targetCurrency]
	if !ok {
		return nil, fmt.Errorf("could not find exchange rate for %s", targetCurrency)
	}

	convertedExchangeRate := money.NewFromFloat(exchangeRate)

	return &convertedExchangeRate, nil
}
