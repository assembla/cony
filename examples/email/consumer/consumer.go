package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/smtp"

	"github.com/assembla/cony"
)

var url = flag.String("url", "amqp://guest:guest@localhost/", "amqp url")

func showUsageAndStatus() {
	fmt.Printf("Consumer is running\n\n")
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Printf("\n\n")
}

// Item struct is used to send down a channel
// the map received from a delivery and the callbacks
// for acknowledge(Ack) or negatively acknowledge(Nack)
type item struct {
	ipt  map[string]string
	ack  func(bool) error
	nack func(bool, bool) error
}

func main() {
	flag.Parse()

	showUsageAndStatus()

	// Channel used for stopping goroutines
	done := make(chan struct{})

	// Items channel used for sending our delivieries to the workers
	itms := make(chan item)

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
		Name:       "email",
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
	)
	cli.Consume(cns)

	// Go routing that uses the cony loop to receive delivieris
	// handle reconnects, etc
	// We use the done channel to exit from this goroutine.
	go func() {
		for cli.Loop() {
			select {
			case msg := <-cns.Deliveries():
				var ipt map[string]string
				err := json.Unmarshal(msg.Body, &ipt)
				if err != nil {
					msg.Reject(false)
				}
				log.Printf("Received body: %q\n", msg.Body)

				// If when we built the consumer we didn't use
				// the "cony.AutoAck()" option this is where we'd
				// have to call the "amqp.Deliveries" methods "Ack",
				// "Nack", "Reject"
				//
				// msg.Ack(false)
				// msg.Nack(false)
				// msg.Reject(false)
				itms <- item{
					ipt:  ipt,
					ack:  msg.Ack,
					nack: msg.Nack,
				}

			case err := <-cns.Errors():
				fmt.Printf("Consumer error: %v\n", err)
			case err := <-cli.Errors():
				fmt.Printf("Client error: %v\n", err)
			case <-done:
				return
			}
		}
	}()

	// Workers
	go func() {
		num := 3
		for i := 0; i < num; i++ {
			go func() {
				for {
					select {
					case itm := <-itms:
						err := sendMail(itm.ipt)
						if err != nil {
							fmt.Println(err)
						}
						itm.ack(false)
					case <-done:
						return
					}
				}
			}()
		}
	}()

	// Block this indefinitely
	<-done
}

func sendMail(ipt map[string]string) error {
	// mail settings
	var (
		usr string = ""
		pwd string = ""
		hst string = "smtp.gmail.com"
		prt string = "587"
	)
	return smtp.SendMail(
		fmt.Sprintf("%v:%v", hst, prt),
		smtp.PlainAuth(
			"",
			usr,
			pwd,
			hst,
		),
		usr,
		[]string{ipt["to"]},
		[]byte(ipt["message"]),
	)
}
