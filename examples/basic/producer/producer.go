package main

import (
	"flag"
	"fmt"

	"time"

	"github.com/LIVEauctioneers/amqp"
	"github.com/LIVEauctioneers/cony"
)

var url = flag.String("url", "amqp://guest:guest@localhost/", "amqp url")
var body = flag.String("body", "Hello world!", "what should be sent")

func showUsageAndStatus() {
	fmt.Printf("Producer is running\n\n")
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Printf("\n\n")
	fmt.Println("Publishing:")
	fmt.Printf("Body: %q\n", *body)
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

	// Declare the exchange we'll be using
	exc := cony.Exchange{
		Name:       "myExc",
		Kind:       "fanout",
		AutoDelete: true,
	}
	cli.Declare([]cony.Declaration{
		cony.DeclareExchange(exc),
	})

	// Declare and register a publisher
	// with the cony client
	pbl := cony.NewPublisher(exc.Name, "pubSub")
	cli.Publish(pbl)
	// Launch a go routine and publish a message.
	// "Publish" is a blocking method this is why it
	// needs to be called in its own go routine.
	//
	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				fmt.Printf("Client publishing\n")
				err := pbl.Publish(amqp.Publishing{
					Body: []byte(*body),
				})
				if err != nil {
					fmt.Printf("Client publish error: %v\n", err)
				}
			}
		}
	}()

	// Client loop sends out declarations(exchanges, queues, bindings
	// etc) to the AMQP server. It also handles reconnecting.
	for cli.Loop() {
		select {
		case err := <-cli.Errors():
			fmt.Printf("Client error: %v\n", err)
		case blocked := <-cli.Blocking():
			fmt.Printf("Client is blocked %v\n", blocked)
		}
	}
}
