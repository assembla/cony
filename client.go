package cony

import (
	"time"

	"github.com/streadway/amqp"
)

// ClientOpt is a Client's functional option type
type ClientOpt func(*Client)

// Client is a Main AMQP client wrapper
type Client struct {
	addr         string
	declarations []Declaration
	consumers    map[*Consumer]struct{}
	publishers   map[*Publisher]struct{}
	errs         chan error
	run          bool
	conn         *amqp.Connection
	bo           Backoffer
	attempt      int
}

// Declare used to declare queues/exchanges/bindings.
// Declaration is saved and will be re-run every time Client gets connection
func (c *Client) Declare(d []Declaration) {
	c.declarations = append(c.declarations, d...)
}

// Consume used to declare consumers
func (c *Client) Consume(cons *Consumer) {
	c.consumers[cons] = struct{}{}
}

func (c *Client) deleteConsumer(cons *Consumer) {
	delete(c.consumers, cons)
}

// Publish used to declare publishers
func (c *Client) Publish(pub *Publisher) {
	c.publishers[pub] = struct{}{}
}

func (c *Client) deletePublisher(pub *Publisher) {
	delete(c.publishers, pub)
}

// Errors returns AMQP connection level errors
func (c *Client) Errors() <-chan error {
	return c.errs
}

// Close shutdown the client
func (c *Client) Close() {
	c.run = false
	c.conn.Close()
	c.conn = nil
}

// Loop should be run as condition for `for` with receiving from (*Client).Errors()
//
// It will manage AMQP connection, run queue and exchange declarations, consumers.
// Will start to return false once (*Client).Close() called.
func (c *Client) Loop() bool {
	var (
		err error
	)

	if !c.run {
		return false
	}

	if c.conn != nil {
		return true
	}

	if c.bo != nil {
		time.Sleep(c.bo.Backoff(c.attempt))
		c.attempt++
	}

	c.conn, err = amqp.Dial(c.addr)
	if c.reportErr(err) {
		return true
	}

	c.attempt = 0

	// guard conn
	go func() {
		chanErr := make(chan *amqp.Error)
		c.conn.NotifyClose(chanErr)

		select {
		case err1 := <-chanErr:
			c.reportErr(err1)
			if c.conn != nil {
				c.conn.Close()
				c.conn = nil
			}
		}
	}()

	ch, err := c.conn.Channel()
	if c.reportErr(err) {
		return true
	}

	for _, declare := range c.declarations {
		c.reportErr(declare(ch))
	}

	for cons := range c.consumers {
		go cons.serve(c)
	}

	for pub := range c.publishers {
		go pub.serve(c)
	}

	return true
}

func (c *Client) reportErr(err error) bool {
	if err != nil {
		select {
		case c.errs <- err:
		default:
		}
		return true
	}
	return false
}

// NewClient initializes new Client
func NewClient(opts ...ClientOpt) *Client {
	c := &Client{
		run:          true,
		declarations: make([]Declaration, 0),
		consumers:    make(map[*Consumer]struct{}),
		publishers:   make(map[*Publisher]struct{}),
		errs:         make(chan error, 100),
	}

	for _, o := range opts {
		o(c)
	}
	return c
}

// URL is a functional option, used in `NewClient` constructor
// default URL is amqp://guest:guest@localhost/
func URL(addr string) ClientOpt {
	return func(c *Client) {
		if addr == "" {
			addr = "amqp://guest:guest@localhost/"
		}
		c.addr = addr
	}
}

// Backoff is a functional option, used to define backoff policy
func Backoff(bo Backoffer) ClientOpt {
	return func(c *Client) {
		c.bo = bo
	}
}
