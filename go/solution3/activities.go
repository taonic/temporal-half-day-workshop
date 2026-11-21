package solution3

import (
	"context"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// The banking Activities are mocked. In a real system each one would call a
// banking API; here they log and succeed, except for one bad account number.

func Withdraw(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Withdrawing", "account", account, "amount", amount)
	return nil
}

func Deposit(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Depositing", "account", account, "amount", amount)
	if account == InvalidAccount {
		// Retrying cannot fix a wrong account number, so tell Temporal not to.
		return temporal.NewNonRetryableApplicationError(
			"deposit failed: account "+account+" does not exist",
			ErrTypeInvalidAccount,
			nil,
		)
	}
	return nil
}

func Refund(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Refunding", "account", account, "amount", amount)
	return nil
}
