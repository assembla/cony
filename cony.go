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
