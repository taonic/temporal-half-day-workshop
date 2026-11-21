package main

import (
	"log"

	"github.com/taonic/temporal-half-day-workshop/go/lab1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, lab1.TaskQueue, worker.Options{})

	// region: register
	// Page 1a registers the Workflow and the Activity.
	// endregion

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
