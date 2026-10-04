package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/VladPetriv/finance_bot/pkg/worker"
	"github.com/google/uuid"
)

type automaticReportEngine struct {
	logger *logger.Logger
	stores Stores
	apis   APIs

	reportGenerationInterval time.Duration
}

// NewAutomaticReportEngine creates a new instance of automaticReportEngine.
func NewAutomaticReportEngine(config *config.Config, logger *logger.Logger, stores Stores, apis APIs) *automaticReportEngine {
	return &automaticReportEngine{
		logger:                   logger,
		stores:                   stores,
		apis:                     apis,
		reportGenerationInterval: config.App.AutomaticReportGenerationInterval,
	}
}

func (a *automaticReportEngine) GenerateReports(ctx context.Context) {
	logger := a.logger.With().Str("name", "automaticReportEngine.GenerateReports").Logger()

	ticker := time.NewTicker(a.reportGenerationInterval)
	defer ticker.Stop()

	pool := worker.NewPool(5, a.generateReport)
	pool.Start(ctx)
	defer pool.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("finished generating automatic reports")
			return
		case <-ticker.C:
			scheduledReportExecutions, err := a.stores.AutomaticReport.ListScheduledReportExecutions(ctx, ListScheduledReportExecutionFilter{
				BetweenFilter: &BetweenFilter{
					From: time.Unix(0, 0).UTC(),
					To:   time.Now().UTC(),
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("list scheduled report executions from store")
				continue
			}
			logger.Debug().Any("scheduledReportExecutions", scheduledReportExecutions).Msg("got scheduled report executions")

			for _, scheduledReportExecution := range scheduledReportExecutions {
				pool.AddJob(scheduledReportExecution.ID, scheduledReportExecution)
			}
		}
	}
}

func (a *automaticReportEngine) generateReport(ctx context.Context, id string, execution model.ScheduledReportExecution) (err error) {
	logger := a.logger.With().Str("name", "automaticReportEngine.generateReport").Logger()
	logger.Debug().Any("execution", execution).Msg("got args")

	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Error().Any("panic", recovered).Msg("recovered from panic")
			err = fmt.Errorf("panic while generating automatic report: %v", recovered)
		}
	}()

	automaticReport, err := a.stores.AutomaticReport.Get(ctx, GetAutomaticReportFilter{
		ID: execution.AutomaticReportID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get automatic report from store")
		return fmt.Errorf("get automatic report from store: %w", err)
	}
	if automaticReport == nil || len(automaticReport.BalanceIDs) == 0 {
		logger.Warn().Msg("automatic report or its balances not found, removing scheduled execution")

		err := a.stores.AutomaticReport.DeleteScheduledReportExecution(ctx, execution.ID)
		if err != nil {
			logger.Error().Err(err).Msg("delete scheduled report execution")
			return fmt.Errorf("delete scheduled report execution: %w", err)
		}

		return ErrAutomaticReportNotFound
	}

	user, err := a.stores.User.Get(ctx, GetUserFilter{
		BalanceID: automaticReport.BalanceIDs[0],
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Warn().Msg("user during automatic report generation not found")
		return ErrUserNotFound
	}

	userCategories, err := a.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories from store")
		return fmt.Errorf("list categories from store: %w", err)
	}

	categories := slices.DeleteFunc(userCategories, func(category model.Category) bool {
		return !slices.Contains(automaticReport.CategoryIDs, category.ID)
	})

	from := automaticReport.Period.SubtractFrom(execution.ExecutionDate)
	for _, balanceID := range automaticReport.BalanceIDs {
		err := a.sendBalanceReport(ctx, sendBalanceReportOptions{
			chatID:          user.ChatID,
			balanceID:       balanceID,
			automaticReport: automaticReport,
			categories:      categories,
			from:            from,
			to:              execution.ExecutionDate,
		})
		if err != nil {
			logger.Error().Err(err).Msg("send balance report")
			return fmt.Errorf("send balance report: %w", err)
		}
	}

	err = a.stores.AutomaticReport.CreateScheduledReportExecution(ctx, &model.ScheduledReportExecution{
		ID:                uuid.NewString(),
		AutomaticReportID: automaticReport.ID,
		ExecutionDate:     automaticReport.Period.CalculateNextExecutionDate(execution.ExecutionDate, time.Now().UTC()),
	})
	if err != nil {
		logger.Error().Err(err).Msg("create next scheduled report execution")
		return fmt.Errorf("create next scheduled report execution: %w", err)
	}

	err = a.stores.AutomaticReport.DeleteScheduledReportExecution(ctx, execution.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete scheduled report execution")
		return fmt.Errorf("delete scheduled report execution: %w", err)
	}

	return nil
}

type sendBalanceReportOptions struct {
	chatID          int
	balanceID       string
	automaticReport *model.AutomaticReport
	categories      []model.Category
	from            time.Time
	to              time.Time
}

func (a *automaticReportEngine) sendBalanceReport(ctx context.Context, opts sendBalanceReportOptions) error {
	logger := a.logger.With().Str("name", "automaticReportEngine.sendBalanceReport").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance, err := a.stores.Balance.Get(ctx, GetBalanceFilter{
		BalanceID:       opts.balanceID,
		PreloadCurrency: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		logger.Warn().Str("balanceID", opts.balanceID).Msg("balance during automatic report generation not found")
		return nil
	}

	operations, err := a.stores.Operation.List(ctx, ListOperationsFilter{
		BalanceID: balance.ID,
		BetweenFilter: &BetweenFilter{
			From: opts.from,
			To:   opts.to,
		},
	})
	if err != nil {
		logger.Error().Err(err).Msg("list operations from store")
		return fmt.Errorf("list operations from store: %w", err)
	}

	operations = slices.DeleteFunc(operations, func(operation model.Operation) bool {
		return !slices.Contains(opts.automaticReport.CategoryIDs, operation.CategoryID)
	})

	statistics, err := model.
		NewStatisticsMessageBuilder(balance, operations, opts.categories).
		BuildForRange(opts.from, opts.to)
	if err != nil {
		logger.Error().Err(err).Msg("build statistics message")
		return fmt.Errorf("build statistics message: %w", err)
	}

	err = a.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.chatID,
		FormatMessageInMarkDown: true,
		Message:                 fmt.Sprintf("🗓 Automatic report: *%s*\n\n%s", opts.automaticReport.Name, statistics),
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message to user")
		return fmt.Errorf("send message to user: %w", err)
	}

	return nil
}
