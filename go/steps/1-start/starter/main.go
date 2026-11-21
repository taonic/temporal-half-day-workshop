package main

// region: imports
import (
	"log"

	"go.temporal.io/sdk/client"
)

// endregion

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// region: start-workflow
	// Page 1a starts the Workflow.
	// endregion
}
