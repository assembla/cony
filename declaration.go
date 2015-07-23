package cony

import "github.com/streadway/amqp"

// Declaration is a callback type to declare AMQP queue/exchange/bidning
type Declaration func(*amqp.Channel) error

// DeclareQueue is a way to declare AMQP queue
func DeclareQueue(q *Queue) Declaration {
	return func(c *amqp.Channel) error {
		realQ, err := c.QueueDeclare(q.Name,
			q.Durable,
			q.AutoDelete,
			q.Exclusive,
			false,
			nil,
		)
		q.l.Lock()
		q.Name = realQ.Name
		q.l.Unlock()
		return err
	}
}

// DeclareExchange is a way to declare AMQP exchange
func DeclareExchange(e Exchange) Declaration {
	return func(c *amqp.Channel) error {
		return c.ExchangeDeclare(e.Name,
			e.Kind,
			e.Durable,
			e.AutoDelete,
			false,
			false,
			nil,
		)
	}
}

// DeclareBinding is a way to declare AMQP bidning between AMQP queue and exchange
func DeclareBinding(b Binding) Declaration {
	return func(c *amqp.Channel) error {
		return c.QueueBind(b.Queue.Name,
			b.Key,
			b.Exchange.Name,
			false,
			nil,
		)
	}
}
