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

type PlaceOrderResult struct {
	Success bool
	Message string
}

func PlaceOrder(ctx workflow.Context, in PlaceOrderInput) (PlaceOrderResult, error) {
	logger := workflow.GetLogger(ctx)

	logger.Info("Placing Order", "Input", in)

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 2.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    5,
			NonRetryableErrorTypes: []string{
				"ErrUnknownPizza",
				"ErrUnknownCustomer",
			},
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var a *Activities

	var menu Menu

	// Retrieve the menu
	if err := workflow.
		ExecuteActivity(ctx, a.RetrieveMenu).
		Get(ctx, &menu); err != nil {
		return PlaceOrderResult{}, err
	}

	logger.Info("Menu retrieved",
		"Menu", menu,
	)

	// Get the requested pizza from the menu
	pizza, err := menu.Pizza(in.Pizza)
	if err != nil {
		message := fmt.Sprintf(
			"You requested pizza %d, which is not on the menu.",
			in.Pizza,
		)

		logger.Info("Pizza not found",
			"Message", message,
		)

		return PlaceOrderResult{
			Success: false,
			Message: message,
		}, nil
	}

	logger.Info("Pizza found in the menu",
		"Pizza", pizza,
	)

	var customer Customer

	// Lookup the customer based on the provided name
	if err := workflow.
		ExecuteActivity(ctx, a.LookupCustomer, in.Name).
		Get(ctx, &customer); err != nil {
		return PlaceOrderResult{}, err
	}

	if customer.Unknown() {
		return PlaceOrderResult{
			Success: false,
			Message: "You are not a known customer",
		}, nil
	}

	logger.Info("Customer found",
		"in.Name", in.Name,
		"Customer", customer,
	)

	order := Order{
		Customer: customer,
		Pizza:    pizza,
	}

	logger.Info("Order placed",
		"Order", order,
	)

	return PlaceOrderResult{
		Success: true,
		Message: fmt.Sprintf(
			"OK, expected delivery of a %s to %s in %s",
			order.Pizza.Name,
			order.Address,
			order.Delivery,
		),
	}, nil
}
