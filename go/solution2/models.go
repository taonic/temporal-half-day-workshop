package solution2

import "time"

const (
	TaskQueue  = "money-transfer-task-queue"
	WorkflowID = "money-transfer-workflow"

	QueryGetStatus = "getStatus"
	SignalApprove  = "approve"

	// ApprovalTimeout is how long the transfer waits for a human decision.
	ApprovalTimeout = 30 * time.Second
)

type TransferRequest struct {
	FromAccount string  `json:"fromAccount"`
	ToAccount   string  `json:"toAccount"`
	Amount      float64 `json:"amount"`
	TransferID  string  `json:"transferId"`
}

type TransferStatus string

const (
	StatusPending          TransferStatus = "PENDING"
	StatusAwaitingApproval TransferStatus = "AWAITING_APPROVAL"
	StatusApproved         TransferStatus = "APPROVED"
	StatusCompleted        TransferStatus = "COMPLETED"
	StatusCancelled        TransferStatus = "CANCELLED"
	StatusTimedOut         TransferStatus = "TIMED_OUT"
)
