package main

import (
	"log"

	"github.com/taonic/temporal-half-day-workshop/go/solution1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, solution1.TaskQueue, worker.Options{})
	w.RegisterWorkflow(solution1.Greeting)
	w.RegisterActivity(solution1.ComposeGreeting)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
