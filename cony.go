package cony

type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
}

type Exchange struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
}

// Used to declare bidning between Queue and Exchange
type Binding struct {
	Queue    *Queue
	Exchange Exchange
	Key      string
}
