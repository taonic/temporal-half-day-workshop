package lab2

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// The banking Activities are mocked. In a real system each one would call a
// banking API; here they log and succeed, so every run is predictable.

func Withdraw(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Withdrawing", "account", account, "amount", amount)
	return nil
}

func Deposit(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Depositing", "account", account, "amount", amount)
	return nil
}

func Refund(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Refunding", "account", account, "amount", amount)
	return nil
}
