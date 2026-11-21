package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taonic/temporal-half-day-workshop/go/lab2"
	"go.temporal.io/sdk/client"
)

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
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, lab2.MoneyTransfer, req)
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	fmt.Println("Started", we.GetID(), "run", we.GetRunID())
	fmt.Printf("Waiting up to %s for an approve Signal...\n", lab2.ApprovalTimeout)

	var status lab2.TransferStatus
	if err := we.Get(context.Background(), &status); err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Println("Final status:", status)
}
