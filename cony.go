// Cony is a high-level wrapper around http://github.com/streadway/amqp library,
// for working declaratively with AMQP. Cony will manage AMQP
// connect/reconnect to AMQP brocker, along with recovery of consumers.

package cony

// Queue hold definition of AMQP queue
type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
}

// Exchange hold definition of AMQP exchange
type Exchange struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
}

// Binding used to declare bidning between AMQP Queue and AMQP Exchange
type Binding struct {
	Queue    *Queue
	Exchange Exchange
	Key      string
}
