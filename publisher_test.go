package cony

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestPublisherImplements_io_Writer(t *testing.T) {
	var _ io.Writer = &Publisher{}
}

func TestPublisher_Cancel(t *testing.T) {
	var stopped bool
	done := make(chan bool)

	p := newTestPublisher()

	go func() {
		defer func() {
			done <- true
		}()
		select {
		case <-p.stop:
			stopped = true
		case <-time.After(1 * time.Millisecond):
			return
		}
	}()
	p.Cancel()
	<-done

	if !stopped {
		t.Error("Cancel() should send stop signal")
	}
}

func TestPublisher_Cancel_willNotBlock(t *testing.T) {
	var ok bool
	p := newTestPublisher()

	go func() {
		p.Cancel()
		p.Cancel()
		p.Cancel()
		ok = true
	}()

	time.Sleep(1 * time.Microsecond) // let goroutine to work
	if !ok {
		t.Error("shold not block")
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
	return NewPublisher("", "", opts...)
}
