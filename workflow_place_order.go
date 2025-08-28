package pds

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type PlaceOrderInput struct {
	Name  string
	Pizza int
}

func PlaceOrder(ctx workflow.Context, in PlaceOrderInput) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        10 * time.Second,
			MaximumAttempts:        5,
			NonRetryableErrorTypes: []string{"UnknownPizza"},
		},
	})

	var a *Activities

	var menu Menu

	if err := workflow.ExecuteActivity(ctx, a.RetrieveMenu).Get(ctx, &menu); err != nil {
		return "", err
	}

	if !menu.HasPizza(in.Pizza) {
		return "", fmt.Errorf("UnknownPizza: You requested pizza %d, which is not on the menu", in.Pizza)
	}

	return "OK", nil
}
