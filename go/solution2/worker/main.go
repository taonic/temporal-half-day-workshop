package main

import (
	"log"

	"github.com/taonic/temporal-half-day-workshop/go/solution2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, solution2.TaskQueue, worker.Options{})
	w.RegisterWorkflow(solution2.MoneyTransfer)
	w.RegisterActivity(solution2.Withdraw)
	w.RegisterActivity(solution2.Deposit)
	w.RegisterActivity(solution2.Refund)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
