package cony

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestAutoAck(t *testing.T) {
	c := newTestConsumer(AutoAck())

	if !c.autoAck {
		t.Error("autoAck shoud be true")
	}
}

func TestAutoTag(t *testing.T) {
	c := newTestConsumer(AutoTag())

	if c.tag == "" {
		t.Error("tag should be set")
	}
}

func TestConsumer_Cancel(t *testing.T) {
	var stopped bool
	done := make(chan bool)

	c := newTestConsumer()

	go func() {
		defer func() {
			done <- true
		}()
		select {
		case <-c.stop:
			stopped = true
		case <-time.After(1 * time.Millisecond):
			return
		}
	}()
	c.Cancel()
	<-done

	if !stopped {
		t.Error("Cancel() should send stop signal")
	}
}

func TestConsumer_Cancel_willNotBlock(t *testing.T) {
	var ok bool
	c := newTestConsumer()

	go func() {
		c.Cancel()
		ok = true
	}()

	time.Sleep(1 * time.Microsecond) // let goroutine to work
	if !ok {
		t.Error("shold not block")
	}
}

func TestConsumer_Deliveries(t *testing.T) {
	c := newTestConsumer()
	go func() {
		select {
		case c.deliveries <- amqp.Delivery{Body: []byte("hello")}:
		case <-time.After(1 * time.Millisecond):
			t.Error("timeout")
		}
	}()

	select {
	case d := <-c.Deliveries():
		if bytes.Compare(d.Body, []byte("hello")) != 0 {
			t.Error("amqp.Delivery should be the same")
		}
	case <-time.After(1 * time.Millisecond):
		t.Error("Deliveries() channel should deliver amqp.Delivery{}")
	}
}

func TestConsumer_Errors(t *testing.T) {
	c := newTestConsumer()
	go func() {
		select {
		case c.errs <- errors.New("Hello error"):
		case <-time.After(1 * time.Millisecond):
			t.Error("timeout")
		}
	}()

	select {
	case err := <-c.Errors():
		if err.Error() != "Hello error" {
			t.Error("Error message should match")
		}
	case <-time.After(1 * time.Millisecond):
		t.Error("Errors() channel should deliver errors")
	}
}

func TestExclusive(t *testing.T) {
	c := newTestConsumer(Exclusive())

	if !c.exclusive {
		t.Error("exclusive should be set")
	}
}

func TestNewConsumer(t *testing.T) {
	var called bool

	NewConsumer(&Queue{}, func(*Consumer) {
		called = true
	})

	if !called {
		t.Error("NewConsumer should call input functional options")
	}
}

func TestNoLocal(t *testing.T) {
	c := newTestConsumer(NoLocal())

	if !c.noLocal {
		t.Error("noLocal should be set")
	}
}

func TestQos(t *testing.T) {
	c := newTestConsumer(Qos(10))

	if c.qos != 10 {
		t.Error("qos should be set to 10")
	}
}

func TestTag(t *testing.T) {
	c := newTestConsumer(Tag("hello"))

	if c.tag != "hello" {
		t.Error("tag should be set to `hello`")
	}
}

func newTestConsumer(opts ...ConsumerOpt) *Consumer {
	q := &Queue{}
	return NewConsumer(q, opts...)
}
