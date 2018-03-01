package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/LIVEauctioneers/cony"
)

var url = flag.String("url", "amqp://guest:guest@localhost/", "amqp url")

func showUsageAndStatus() {
	fmt.Printf("Consumer is running\n\n")
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Printf("\n\n")
}

func main() {
	flag.Parse()

	showUsageAndStatus()

	// Construct new client with the flag url
	// and default backoff policy
	cli := cony.NewClient(
		cony.URL(*url),
		cony.Backoff(cony.DefaultBackoff),
	)

	// Declarations
	// The queue name will be supplied by the AMQP server
	que := &cony.Queue{
		AutoDelete: true,
	}
	exc := cony.Exchange{
		Name:       "web",
		Kind:       "fanout",
		AutoDelete: true,
	}
	bnd := cony.Binding{
		Queue:    que,
		Exchange: exc,
		Key:      "",
	}
	cli.Declare([]cony.Declaration{
		cony.DeclareQueue(que),
		cony.DeclareExchange(exc),
		cony.DeclareBinding(bnd),
	})

	// Declare and register a consumer
	cns := cony.NewConsumer(
		que,
		cony.AutoAck(), // Auto sign the deliveries
	)
	cli.Consume(cns)
	for cli.Loop() {
		select {
		case msg := <-cns.Deliveries():
			log.Printf("Received body: %q\n", msg.Body)
			// If when we built the consumer we didn't use
			// the "cony.AutoAck()" option this is where we'd
			// have to call the "amqp.Deliveries" methods "Ack",
			// "Nack", "Reject"
			//
			// msg.Ack(false)
			// msg.Nack(false)
			// msg.Reject(false)
		case err := <-cns.Errors():
			fmt.Printf("Consumer error: %v\n", err)
		case err := <-cli.Errors():
			fmt.Printf("Client error: %v\n", err)
		}
	}
}
