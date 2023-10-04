package cony

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestNewClient(t *testing.T) {
	c := NewClient(func(c *Client) {
		c.addr = "test1"
	})

	if c.addr != "test1" {
		t.Error("should call functional opts")
	}
}

func TestClient_Declare(t *testing.T) {
	c := &Client{}

	if len(c.declarations) != 0 {
		t.Error("declarations should be empty")
	}

	q := &Queue{}
	c.Declare([]Declaration{DeclareQueue(q)})

	if len(c.declarations) != 1 {
		t.Error("declarations should have 1 declaration")
	}
}

func TestClient_Consume(t *testing.T) {
	c := NewClient()
	cons := &Consumer{}
	c.Consume(cons)

	if _, ok := c.consumers[cons]; !ok {
		t.Error("should save consumer")
	}

	c.deleteConsumer(cons)

	if _, ok := c.consumers[cons]; ok {
		t.Error("should remove consumer")
	}
}

func TestClient_Publish(t *testing.T) {
	c := NewClient()
	pub := &Publisher{}
	c.Publish(pub)

	if _, ok := c.publishers[pub]; !ok {
		t.Error("should save publisher")
	}

	c.deletePublisher(pub)

	if _, ok := c.publishers[pub]; ok {
		t.Error("should remove publisher")
	}
}

func TestClient_Errors(t *testing.T) {
	c := NewClient()
	errs := c.Errors()

	if errs == nil {
		t.Error("should initialize Errors channel")
	}
}

func TestClient_Close(t *testing.T) {
	c := NewClient()
	c.Close()

	if c.run != noRun {
		t.Error("should stop running")
	}
}

func TestClient_Loop(t *testing.T) {
	c := NewClient()
	c.run = noRun

	if c.Loop() {
		t.Error("should not run if noRun")
	}

	// immitate connection
	c.conn.Store(&amqp.Connection{})
	c.run = run

	if !c.Loop() {
		t.Error("should tell loop to run")
	}
}

func TestClient_reportErr(t *testing.T) {
	c := NewClient()

	if c.reportErr(nil) {
		t.Error("should return false on no error")
	}

	// fill in errs buffer
	for i := 0; i < 100; i++ {
		c.reportErr(errors.New("test err"))
	}

	// should not block, error will be discarded
	if !c.reportErr(errors.New("test err")) {
		t.Error("should return true")
	}
}

func TestClient_channel(t *testing.T) {}

func TestClient_connection(t *testing.T) {
	c := NewClient()

	if _, err := c.connection(); err != ErrNoConnection {
		t.Error("error should be", ErrNoConnection)
	}

	c.conn.Store(&amqp.Connection{})

	con, err := c.connection()
	if con == nil {
		t.Error("should return existing connection")
	}

	if err != nil {
		t.Error("should be no errors")
	}
}

func TestURL(t *testing.T) {
	c := &Client{}
	URL("test1")(c)

	if c.addr != "test1" {
		t.Error("should set URL")
	}

	URL("")(c)
	if c.addr != "amqp://guest:guest@localhost/" {
		t.Error("should use default URL")
	}
}

func TestBackoff(t *testing.T) {
	c := &Client{}
	Backoff(DefaultBackoff)(c)

	if c.bo == nil {
		t.Error("should set backoff")
	}
}
