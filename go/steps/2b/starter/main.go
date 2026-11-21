package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taonic/temporal-half-day-workshop/go/lab2"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// Starts a transfer and exits. The transfer keeps running on the server:
// approve it, query it and follow it from the CLI.
func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	req := lab2.TransferRequest{
		FromAccount: "account-123",
		ToAccount:   "account-456",
		Amount:      100,
		TransferID:  "transfer-1",
	}
	options := client.StartWorkflowOptions{
		ID:        lab2.WorkflowID,
		TaskQueue: lab2.TaskQueue,
		// Running the starter again replaces a transfer that is still waiting.
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING,
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, lab2.MoneyTransfer, req)
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	fmt.Println("Started", we.GetID())
}
