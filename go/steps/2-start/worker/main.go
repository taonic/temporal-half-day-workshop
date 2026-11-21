package main

import (
	"log"

	"github.com/taonic/temporal-half-day-workshop/go/lab2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, lab2.TaskQueue, worker.Options{})
	w.RegisterWorkflow(lab2.MoneyTransfer)
	w.RegisterActivity(lab2.Withdraw)
	w.RegisterActivity(lab2.Deposit)
	w.RegisterActivity(lab2.Refund)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
