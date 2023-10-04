package cony

import (
	"bytes"
	"errors"
	"io"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestPublisherImplements_io_Writer(t *testing.T) {
	var _ io.Writer = &Publisher{}
}

func TestPublisher_Cancel(t *testing.T) {
	var (
		stopped bool
		runSync = make(chan bool)
	)

	p := newTestPublisher()

	go func() {
		<-runSync
		<-p.stop
		stopped = true
		runSync <- true
	}()

	runSync <- true
	p.Cancel()

	// just make sure that multiple calls will not blow up
	p.Cancel()
	p.Cancel()
	<-runSync

	if !stopped {
		t.Error("Cancel() should send stop signal")
	}
}

func TestPublisher_Cancel_willNotBlock(t *testing.T) {
	var (
		ok      bool
		runSync = make(chan bool)
	)
	p := newTestPublisher()

	go func() {
		<-runSync
		p.Cancel()
		p.Cancel()
		p.Cancel()
		ok = true
		runSync <- true
	}()

	runSync <- true
	<-runSync
	if !ok {
		t.Error("shold not block")
	}
}

func TestPublisher_serve(t *testing.T) {
	var (
		runSync      = make(chan bool)
		deleted      bool
		closed       bool
		notifyClose  bool
		exchangeName string
		routingKey   string
		testMsg      *amqp.Publishing
	)

	p := newTestPublisher()
	cli := &mqDeleterTest{
		_deletePublisher: func(*Publisher) {
			deleted = true
		},
	}

	ch1 := &mqChannelTest{
		_Close: func() error {
			closed = true
			return nil
		},
		_NotifyClose: func(errChan chan *amqp.Error) chan *amqp.Error {
			notifyClose = true
			return errChan
		},
		_Publish: func(ex string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
			exchangeName = ex
			routingKey = key
			testMsg = &msg
			return nil
		},
	}

	go func() {
		<-runSync
		p.serve(cli, ch1)
		runSync <- true
	}()

	runSync <- true
	p.Write([]byte("test1"))
	p.Cancel()
	<-runSync

	if !notifyClose {
		t.Error("should register notifyClose")
	}

	if !deleted {
		t.Error("should delete publisher")
	}

	if !closed {
		t.Error("should close channel")
	}

	if exchangeName != "exchange.name" {
		t.Error("should set correct routing key")
	}

	if routingKey != "routing.key" {
		t.Error("should set correct routing key")
	}

	if bytes.Compare(testMsg.Body, []byte("test1")) != 0 {
		t.Error("should publish correct messaged")
	}
}

func TestPublisher_serve_customRoutingKey(t *testing.T) {
	var (
		runSync      = make(chan bool)
		deleted      bool
		closed       bool
		notifyClose  bool
		exchangeName string
		routingKey   string
		testMsg      *amqp.Publishing
	)

	p := newTestPublisher()
	cli := &mqDeleterTest{
		_deletePublisher: func(*Publisher) {
			deleted = true
		},
	}

	ch1 := &mqChannelTest{
		_Close: func() error {
			closed = true
			return nil
		},
		_NotifyClose: func(errChan chan *amqp.Error) chan *amqp.Error {
			notifyClose = true
			return errChan
		},
		_Publish: func(ex string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
			exchangeName = ex
			routingKey = key
			testMsg = &msg
			return nil
		},
	}

	go func() {
		<-runSync
		p.serve(cli, ch1)
		runSync <- true
	}()

	runSync <- true
	p.PublishWithRoutingKey(amqp.Publishing{Body: []byte("test1")}, "my.routing.key")
	p.Cancel()
	<-runSync

	if !notifyClose {
		t.Error("should register notifyClose")
	}

	if !deleted {
		t.Error("should delete publisher")
	}

	if !closed {
		t.Error("should close channel")
	}

	if exchangeName != "exchange.name" {
		t.Error("should set correct routing key")
	}

	if routingKey != "my.routing.key" {
		t.Error("should set correct routing key")
	}

	if bytes.Compare(testMsg.Body, []byte("test1")) != 0 {
		t.Error("should publish correct messaged")
	}
}

func TestPublisher_serve_errors(t *testing.T) {
	var (
		runSync          = make(chan bool)
		testErrChan      chan *amqp.Error
		testPublishErr   = errors.New("pub err")
		_deletePublisher bool
	)

	p := newTestPublisher()
	cli := &mqDeleterTest{
		_deletePublisher: func(*Publisher) {
			_deletePublisher = true
		},
	}

	ch1 := &mqChannelTest{
		_Close: func() error {
			return nil
		},
		_NotifyClose: func(errChan chan *amqp.Error) chan *amqp.Error {
			testErrChan = errChan
			return errChan
		},
		_Publish: func(string, string, bool, bool, amqp.Publishing) error {
			return testPublishErr
		},
	}

	go func() {
		<-runSync
		p.serve(cli, ch1)
		runSync <- true
	}()

	runSync <- true
	_, err := p.Write([]byte("test1"))
	close(testErrChan) // immitate amqp.Channel close
	<-runSync

	if err != testPublishErr {
		t.Error("should return correct error")
	}

	if _deletePublisher {
		t.Error("on errors should not delete publisher")
	}
}

func TestNewPublisher(t *testing.T) {
	var called bool

	NewPublisher("", "", func(*Publisher) {
		called = true
	})

	if !called {
		t.Error("NewPublisher should call input functional options")
	}
}

func TestPublisher_Publish(t *testing.T) {
	var ok bool
	testBuf := []byte("test1")
	tpl1 := amqp.Publishing{AppId: "app1"}
	msg1 := amqp.Publishing{AppId: "app2", Body: testBuf}
	p := newTestPublisher(PublishingTemplate(tpl1))
	testErr := errors.New("testing error")

	go func() {
		envelop := <-p.pubChan
		msg := <-envelop.pub
		if bytes.Compare(msg.Body, testBuf) == 0 {
			if msg.AppId == "app2" {
				ok = true
			}
		}
		envelop.err <- testErr
	}()

	err := p.Publish(msg1)
	if err != testErr {
		t.Error("Publish should receive correct error")
	}

	if !ok {
		t.Error("Publish() sent wrong message")
	}
}

func TestPublisher_PublishWithStop(t *testing.T) {
	msg1 := amqp.Publishing{Body: []byte("test1")}
	p := newTestPublisher()

	go func() {
		p.Cancel()
	}()

	err := p.Publish(msg1)
	if err != ErrPublisherDead {
		t.Error("Publish should receive", ErrPublisherDead)
	}
}

func TestPublisher_Write(t *testing.T) {
	var ok bool
	testBuf := []byte("test1")
	p := newTestPublisher(PublishingTemplate(amqp.Publishing{AppId: "app1"}))
	testErr := errors.New("testing error")

	go func() {
		envelop := <-p.pubChan
		msg := <-envelop.pub
		if bytes.Compare(msg.Body, testBuf) == 0 {
			if msg.AppId == "app1" {
				ok = true
			}
		}
		envelop.err <- testErr
	}()

	n, err := p.Write(testBuf)

	if n != len(testBuf) {
		t.Error("Write() should return len equal to input buffer")
	}

	if err != testErr {
		t.Error("Write() should return correct error")
	}

	if !ok {
		t.Error("Write() send incorrect message")
	}
}

func TestPublishingTemplate(t *testing.T) {
	p := newTestPublisher()
	pub := amqp.Publishing{AppId: "ololoapp"}

	PublishingTemplate(pub)(p)

	if p.tmpl.AppId != "ololoapp" {
		t.Error("PublishingTemplate() should update template")
	}
}

func newTestPublisher(opts ...PublisherOpt) *Publisher {
	return NewPublisher("exchange.name", "routing.key", opts...)
}
