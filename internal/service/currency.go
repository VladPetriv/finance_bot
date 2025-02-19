package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

type currencyService struct {
	logger   *logger.Logger
	storages Stores
	apis     APIs
}

// NewCurrency returns new instance of currency service.
func NewCurrency(logger *logger.Logger, apis APIs, storages Stores) *currencyService {
	return &currencyService{
		logger:   logger,
		apis:     apis,
		storages: storages,
	}
}

// availableCurrenciesCount represents the number of available currencies which can be received from the CurrencyExchanger API.
const availableCurrenciesCount = 161

func (c *currencyService) InitCurrencies(ctx context.Context) error {
	logger := c.logger.With().Str("name", "currencyService.InitCurrencies").Logger()
	logger.Debug().Msg("got args")

	currenciesCount, err := c.storages.Currency.Count(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("count currencies in the database")
		return fmt.Errorf("count currencies in the database: %w", err)
	}
	if currenciesCount == availableCurrenciesCount {
		logger.Info().Any("currenciesCount", currenciesCount).Msg("currencies already initialized")
		return nil
	}

	currencies, err := c.apis.CurrencyExchanger.FetchCurrencies()
	if err != nil {
		logger.Error().Err(err).Msg("fetch currencies through currency exchanger")
		return fmt.Errorf("fetch currencies through currency exchanger: %w", err)
	}
	if len(currencies) == 0 {
		logger.Info().Msg("currencies not found")
		return nil
	}

	for _, currency := range currencies {
		err := c.storages.Currency.CreateIfNotExists(ctx, &models.Currency{
			ID:     uuid.NewString(),
			Name:   currency.Name,
			Code:   currency.Code,
			Symbol: currency.Symbol,
		})
		if err != nil {
			logger.Error().Err(err).Any("currency", currency).Msg("create currency in the database")

			continue
		}
	}

	return nil
}

func (c *currencyService) Convert(ctx context.Context, opts ConvertCurrencyOptions) (*money.Money, error) {
	logger := c.logger.With().Str("name", "currencyService.Convert").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	exchangeRate, err := c.apis.CurrencyExchanger.GetExchangeRate(opts.BaseCurrency, opts.TargetCurrency)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Msg(err.Error())
			return nil, err
		}

		logger.Error().Err(err).Msg("get exchange rate through currency exchanger")
		return nil, fmt.Errorf("get exchange rate through currency exchanger: %w", err)
	}
	logger.Debug().Any("exchangeRate", exchangeRate).Msg("got exchange rate")

	opts.Amount.Mul(*exchangeRate)

	return &opts.Amount, nil
}
