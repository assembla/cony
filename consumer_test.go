package cony

import (
	"bytes"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
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
		<-done
		select {
		case <-c.stop:
			stopped = true
		case <-time.After(1 * time.Millisecond):
			return
		}
		done <- true
	}()

	done <- true
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
		c.Cancel()
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

func TestConsumer_reportErr(t *testing.T) {
	var (
		okDefault, okNil bool
	)

	c := newTestConsumer()
	testErr := errors.New("test error")

	go func() {
		for i := 0; i <= 101; i++ {
			c.reportErr(testErr)
		}
		okDefault = c.reportErr(testErr)
		okNil = !c.reportErr(nil)
	}()

	err := <-c.errs
	if err != testErr {
		t.Error("error should be the same")
	}

	if !okDefault {
		t.Error("reportErr should not block")
	}

	if !okNil {
		t.Error("reportErr should return false on nil error")
	}
}

func TestConsumer_serve_qos(t *testing.T) {
	var (
		qos      int
		qosError = errors.New("Qos")
		runSync  = make(chan bool)
	)

	c := newTestConsumer(Qos(10))

	ch1 := &mqChannelTest{
		_Qos: func(prefetchCount int, prefetchSize int, global bool) error {
			qos = prefetchCount
			return qosError
		},
	}

	go func() {
		<-runSync
		c.serve(nil, ch1)
		runSync <- true
	}()

	runSync <- true
	err := <-c.errs
	<-runSync

	if qos != 10 {
		t.Error("consumer should declare qos")
	}

	if err != qosError {
		t.Error("reported error should be qos")
	}
}

func TestConsumer_serve_Consume(t *testing.T) {
	var (
		runSync      = make(chan bool)
		consumeError = errors.New("consume error")
	)

	c := newTestConsumer()

	ch1 := &mqChannelTest{
		_Qos: func(int, int, bool) error {
			return nil
		},
		_Consume: func(name string, tag string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
			return nil, consumeError
		},
	}

	go func() {
		<-runSync
		c.serve(nil, ch1)
		runSync <- true
	}()

	runSync <- true
	err := <-c.errs
	<-runSync

	if err != consumeError {
		t.Error("reported error should be consume")
	}
}

func TestConsumer_serve_for(t *testing.T) {
	var (
		runSync    = make(chan bool)
		deleted    bool
		closed     bool
		deliveries = make(chan amqp.Delivery)
	)

	c := newTestConsumer()
	cli := &mqDeleterTest{
		_deleteConsumer: func(*Consumer) {
			deleted = true
		},
	}

	ch1 := &mqChannelTest{
		_Qos: func(int, int, bool) error {
			return nil
		},
		_Consume: func(string, string, bool, bool, bool, bool, amqp.Table) (<-chan amqp.Delivery, error) {
			return deliveries, nil
		},
		_Close: func() error {
			closed = true
			return nil
		},
	}

	go func() {
		<-runSync
		c.serve(cli, ch1)
		runSync <- true
	}()

	runSync <- true

	deliveries <- amqp.Delivery{Body: []byte("test1")}
	msg := <-c.Deliveries()

	go func() {
		c.Cancel()
	}()
	<-runSync

	if bytes.Compare(msg.Body, []byte("test1")) != 0 {
		t.Error("should deliver correct message")
	}

	if !deleted {
		t.Error("should delete consumer")
	}

	if !closed {
		t.Error("should close channel")
	}

	go func() {
		<-runSync
		c.serve(cli, ch1)
		runSync <- true
	}()

	runSync <- true
	close(deliveries) // immitate close of amqp.Channel
	<-runSync
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
