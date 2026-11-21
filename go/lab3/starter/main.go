package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taonic/temporal-half-day-workshop/go/lab3"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// The deposit target is wrong on purpose.
	req := lab3.TransferRequest{
		FromAccount: "account-123",
		ToAccount:   lab3.InvalidAccount,
		Amount:      100,
		TransferID:  "transfer-1",
	}
	options := client.StartWorkflowOptions{
		ID:        lab3.WorkflowID,
		TaskQueue: lab3.TaskQueue,
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, lab3.MoneyTransfer, req)
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	fmt.Println("Started", we.GetID(), "run", we.GetRunID())

	time.Sleep(2 * time.Second)
	if err := c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), lab3.SignalApprove, true); err != nil {
		log.Fatalln("Unable to send approve Signal", err)
	}
	fmt.Println("Approved. Waiting for the result; send updateToAccount if it parks in RETRYING...")

	var status lab3.TransferStatus
	if err := we.Get(context.Background(), &status); err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Println("Final status:", status)
}
