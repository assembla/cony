package cony

import (
	"testing"

	"github.com/streadway/amqp"
)

type testDeclarer struct {
	_QueueDeclare    func() (amqp.Queue, error)
	_ExchangeDeclare func() error
	_QueueBind       func() error
}

func (td *testDeclarer) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return td._QueueDeclare()
}

func (td *testDeclarer) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return td._ExchangeDeclare()
}

func (td *testDeclarer) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return td._QueueBind()
}

func TestDeclareQueue(t *testing.T) {
	var ok bool

	q := &Queue{
		Name: "Q1",
	}

	td := &testDeclarer{
		_QueueDeclare: func() (amqp.Queue, error) {
			ok = true
			return amqp.Queue{Name: "Q1_REAL"}, nil
		},
	}

	DeclareQueue(q)(td)

	if !ok {
		t.Error("DeclareQueue() should call declarer.QueueDeclare()")
	}

	if q.Name != "Q1_REAL" {
		t.Error("DeclareQueue() should update queue name from AMQP reply")
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
