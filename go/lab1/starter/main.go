package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taonic/temporal-half-day-workshop/go/lab1"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	options := client.StartWorkflowOptions{
		ID:        "greeting-workflow",
		TaskQueue: lab1.TaskQueue,
	}

	// TODO: Start the Workflow:
	//       we, err := c.ExecuteWorkflow(context.Background(), options, lab1.Greeting, "Temporal")
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	fmt.Println("Started", we.GetID(), "run", we.GetRunID())

	var greeting string
	if err := we.Get(context.Background(), &greeting); err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Println(greeting)
}
