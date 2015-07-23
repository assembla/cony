package cony

import (
	"fmt"
	"os"

	"github.com/streadway/amqp"
)

// Consumer's functional option type
type ConsumerOpt func(*Consumer)

type Consumer struct {
	q          *Queue
	deliveries chan amqp.Delivery
	errs       chan error
	qos        int
	tag        string
	autoAck    bool
	exclusive  bool
	noLocal    bool
	stop       chan struct{}
}

// Receive deliveries, shipped to this consumer
// this channel never closed, even on disconnects
func (c *Consumer) Deliveries() <-chan amqp.Delivery {
	return c.deliveries
}

// Receive channel level errors
func (c *Consumer) Errors() <-chan error {
	return c.errs
}

// Cancel this consumer
func (c *Consumer) Cancel() {
	select {
	case c.stop <- struct{}{}:
	default:
	}
}

func (c *Consumer) reportErr(err error) bool {
	if err != nil {
		select {
		case c.errs <- err:
		default:
		}
		return true
	}
	return false
}

func (c *Consumer) serve(client *Client) {
	if client.conn == nil {
		return
	}

	ch1, err1 := client.conn.Channel()
	if c.reportErr(err1) {
		return
	}

	if c.reportErr(ch1.Qos(c.qos, 0, false)) {
		return
	}

	deliveries, err2 := ch1.Consume(c.q.Name,
		c.tag,       // consumer tag
		c.autoAck,   // autoAck,
		c.exclusive, // exclusive,
		c.noLocal,   // noLocal,
		false,       // noWait,
		nil,         // args Table
	)
	if c.reportErr(err2) {
		return
	}

	for {
		select {
		case <-c.stop:
			client.deleteConsumer(c)
			ch1.Close()
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			c.deliveries <- d
		}
	}
}

// Consumer's constructor
func NewConsumer(q *Queue, opts ...ConsumerOpt) *Consumer {
	c := &Consumer{
		q:          q,
		deliveries: make(chan amqp.Delivery),
		errs:       make(chan error, 100),
		stop:       make(chan struct{}, 2),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Set Qos on channel
func Qos(count int) ConsumerOpt {
	return func(c *Consumer) {
		c.qos = count
	}
}

// Set tag for consumer
func Tag(tag string) ConsumerOpt {
	return func(c *Consumer) {
		c.tag = tag
	}
}

// Set automatically generated tag like this
//	fmt.Sprintf(QueueName+"-pid-%d@%s", os.Getpid(), os.Hostname())
func AutoTag() ConsumerOpt {
	return func(c *Consumer) {
		host, _ := os.Hostname()
		tag := fmt.Sprintf(c.q.Name+"-pid-%d@%s", os.Getpid(), host)
		Tag(tag)(c)
	}
}

// Set this consumer in AutoAck mode
func AutoAck() ConsumerOpt {
	return func(c *Consumer) {
		c.autoAck = true
	}
}

// Set this consumer in exclusive mode
func Exclusive() ConsumerOpt {
	return func(c *Consumer) {
		c.exclusive = true
	}
}

// Set this consumer in NoLocal mode.
func NoLocal() ConsumerOpt {
	return func(c *Consumer) {
		c.noLocal = true
	}
}
