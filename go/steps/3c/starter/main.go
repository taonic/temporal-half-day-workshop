package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taonic/temporal-half-day-workshop/go/lab3"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// Starts a transfer to an account that does not exist, approves it two
// seconds later, and exits. Follow and fix the transfer from the CLI.
func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	req := lab3.TransferRequest{
		FromAccount: "account-123",
		ToAccount:   lab3.InvalidAccount, // wrong on purpose
		Amount:      100,
		TransferID:  "transfer-1",
	}
	options := client.StartWorkflowOptions{
		ID:        lab3.WorkflowID,
		TaskQueue: lab3.TaskQueue,
		// Running the starter again replaces a transfer that is still waiting.
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING,
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, lab3.MoneyTransfer, req)
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	fmt.Println("Started", we.GetID(), "with a deposit to", req.ToAccount)

	time.Sleep(2 * time.Second)
	if err := c.SignalWorkflow(context.Background(), we.GetID(), "", lab3.SignalApprove, true); err != nil {
		log.Fatalln("Unable to send approve Signal", err)
	}
	fmt.Println("Approved")
}
