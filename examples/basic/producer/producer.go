package main

import (
	"flag"
	"fmt"

	"github.com/assembla/cony"
	"github.com/streadway/amqp"
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
		Name:       "basic",
		Kind:       "fanout",
		AutoDelete: true,
	}
	cli.Declare([]cony.Declaration{
		cony.DeclareExchange(exc),
	})

	// Declare and register a publisher
	// with the cony client
	pbl := cony.NewPublisher(exc.Name, "")
	cli.Publish(pbl)

	// Chan used for signaling when
	// publishing has been completed
	// so the program can exit
	done := make(chan struct{})

	// Launch a go routine and publish a message.
	// "Publish" is a blocking method this is why it
	// needs to be called in its own go routine.
	//
	go func() {
		pbl.Publish(amqp.Publishing{
			Body: []byte(*body),
		})
		// Close done to signal that the program
		// can exit
		close(done)
	}()

	// Client loop sends out declarations(exchanges, queues, bindings
	// etc) to the AMQP server. It also handles reconnecting.
	// We use the done channel here to make sure our publishing is
	// actually published before the programe exists.
	// We then close the client to exit the Loop.
	for cli.Loop() {
		select {
		case err := <-cli.Errors():
			fmt.Printf("Client error: %v\n", err)
		case <-done:
			cli.Close()
		}
	}
}
