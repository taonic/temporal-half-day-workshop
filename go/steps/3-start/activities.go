package lab3

// region: imports
import (
	"context"

	"go.temporal.io/sdk/activity"
)

// endregion

// The banking Activities are mocked. In a real system each one would call a
// banking API; here they log and succeed, except for one bad account number.

func Withdraw(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Withdrawing", "account", account, "amount", amount)
	return nil
}

func Deposit(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Depositing", "account", account, "amount", amount)

	// region: invalid-account
	// Page 3a rejects the bad account here. For now every deposit succeeds.
	// endregion

	return nil
}

func Refund(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Refunding", "account", account, "amount", amount)
	return nil
}
