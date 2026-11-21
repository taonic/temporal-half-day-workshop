package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taonic/temporal-half-day-workshop/go/solution2"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	req := solution2.TransferRequest{
		FromAccount: "account-123",
		ToAccount:   "account-456",
		Amount:      100,
		TransferID:  "transfer-1",
	}
	options := client.StartWorkflowOptions{
		ID:        solution2.WorkflowID,
		TaskQueue: solution2.TaskQueue,
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, solution2.MoneyTransfer, req)
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	fmt.Println("Started", we.GetID(), "run", we.GetRunID())
	fmt.Printf("Waiting up to %s for an approve Signal...\n", solution2.ApprovalTimeout)

	var status solution2.TransferStatus
	if err := we.Get(context.Background(), &status); err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Println("Final status:", status)
}
