package cony

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type testDeclarer struct {
	_QueueDeclare    func(string) (amqp.Queue, error)
	_ExchangeDeclare func() error
	_QueueBind       func() error
}

func (td *testDeclarer) QueueDeclare(name string, durable, autoDelete,
	exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return td._QueueDeclare(name)
}

func (td *testDeclarer) ExchangeDeclare(name, kind string, durable, autoDelete,
	internal, noWait bool, args amqp.Table) error {
	return td._ExchangeDeclare()
}

func (td *testDeclarer) QueueBind(name, key, exchange string, noWait bool,
	args amqp.Table) error {
	return td._QueueBind()
}

func TestDeclareQueue(t *testing.T) {
	var (
		callOK, nameOK bool
	)

	q := &Queue{
		Name: "Q1",
	}

	td := &testDeclarer{
		_QueueDeclare: func(name string) (amqp.Queue, error) {
			callOK = true
			if name == "Q1" {
				nameOK = true
			}
			return amqp.Queue{Name: "Q1_REAL"}, nil
		},
	}

	testDec := DeclareQueue(q)
	testDec(td)

	if !callOK {
		t.Error("DeclareQueue() should call declarer.QueueDeclare()")
	}

	if q.Name != "Q1_REAL" {
		t.Error("DeclareQueue() should update queue name from AMQP reply")
	}

	// call it another time (like reconnect event happened)
	testDec(td)
	if !nameOK {
		t.Error("queue name should be preserved")
	}
}

func TestDeclareExchange(t *testing.T) {
	var ok bool

	e := Exchange{Name: "ex1"}

	td := &testDeclarer{
		_ExchangeDeclare: func() error {
			ok = true
			return nil
		},
	}

	DeclareExchange(e)(td)

	if !ok {
		t.Error("DeclareExchange() should call declarer.ExchangeDeclare()")
	}
}

func TestDeclareBinding(t *testing.T) {
	var ok bool

	b := Binding{
		Queue:    &Queue{Name: "lol1"},
		Exchange: Exchange{Name: "lol2"},
		Key:      "ololoev",
	}

	td := &testDeclarer{
		_QueueBind: func() error {
			ok = true
			return nil
		},
	}

	DeclareBinding(b)(td)

	if !ok {
		t.Error("DeclareBinding() should call declarer.QueueBind()")
	}
}
