package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/assembla/cony"
	amqp "github.com/rabbitmq/amqp091-go"
)

var port = flag.Int("port", 3000, "listening port")
var url = flag.String("url", "amqp://guest:guest@localhost/", "amqp url")

var form = `
	{{ if eq .status "thanks"}}
		<p>Thank you</p>
	{{ end }}
	<form method="post">
		<input type="text" name="body" style="width:300px" />
		<input type="submit" value="Send" />
	</form>
`

func showUsageAndStatus() {
	fmt.Printf("Producer is running\n")
	fmt.Printf("Listening on: %v\n\n", *port)
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

	// Declare the exchange we'll be using
	exc := cony.Exchange{
		Name:       "web",
		Kind:       "fanout",
		AutoDelete: true,
	}
	cli.Declare([]cony.Declaration{
		cony.DeclareExchange(exc),
	})

	// Declare and register a publisher
	// with the cony client.
	// This needs to be "global" per client
	// and we'll need to use this exact value in
	// our handlers (contexts should be of help)
	pbl := cony.NewPublisher(exc.Name, "")
	cli.Publish(pbl)

	// Start our loop in a new gorouting
	// so we don't block this one
	go func() {
		for cli.Loop() {
			select {
			case err := <-cli.Errors():
				fmt.Println(err)
			}
		}
	}()

	// Simple template for our web-view
	tpl, err := template.New("form").Parse(form)
	if err != nil {
		log.Fatal(err)
		return
	}

	// HTTP handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// "GET" shows the template along
			// with the possible thanks message
			hdr := w.Header()
			hdr["Content-Type"] = []string{"text/html"}
			tpl.Execute(w, map[string]string{
				"status": r.FormValue("status"),
			})
			return
		} else if r.Method == "POST" {
			// "POST" publishes the value received
			// from the form to AMQP
			// Note: we're using the "pbl" variable
			// (declared above in our code) and we
			// don't declare a new Publisher value.
			go pbl.Publish(amqp.Publishing{
				Body: []byte(r.FormValue("body")),
			})
			http.Redirect(w, r, "/?status=thanks", http.StatusFound)
			return
		}
		http.Error(w, "404 not found", http.StatusNotFound)
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), nil))
}
