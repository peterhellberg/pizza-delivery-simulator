package pds

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func GetMenu(ctx workflow.Context) (Menu, error) {
	logger := workflow.GetLogger(ctx)

	logger.Info("Getting the menu")

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    5,
		},
	})

	var a *Activities

	var menu Menu

	if err := workflow.
		ExecuteActivity(ctx, a.RetrieveMenu).
		Get(ctx, &menu); err != nil {
		if errors.Is(err, ErrUnavailableMenu) {
			logger.Error("Menu is currently unavailable")
		}

		return menu, err
	}

	return menu, nil
}
