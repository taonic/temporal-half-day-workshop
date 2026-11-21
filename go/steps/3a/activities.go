package lab3

// region: imports
import (
	"context"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
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
	if account == InvalidAccount {
		return temporal.NewNonRetryableApplicationError(
			"deposit failed: account "+account+" does not exist",
			ErrTypeInvalidAccount,
			nil,
		)
	}
	// endregion

	return nil
}

func Refund(ctx context.Context, account string, amount float64) error {
	activity.GetLogger(ctx).Info("Refunding", "account", account, "amount", amount)
	return nil
}
