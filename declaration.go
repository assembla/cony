package cony

import "github.com/streadway/amqp"

type Declaration func(*amqp.Channel) error

func DeclareQueue(q *Queue) Declaration {
	return func(c *amqp.Channel) error {
		realQ, err := c.QueueDeclare(q.Name,
			q.Durable,
			q.AutoDelete,
			q.Exclusive,
			false,
			nil,
		)
		q.Name = realQ.Name
		return err
	}
}

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
