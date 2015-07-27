// Package cony is a high-level wrapper around http://github.com/streadway/amqp library,
// for working declaratively with AMQP. Cony will manage AMQP
// connect/reconnect to AMQP brocker, along with recovery of consumers.
package cony

import (
	"sync"

	"github.com/streadway/amqp"
)

// Queue hold definition of AMQP queue
type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool

	l sync.Mutex
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

type mqDeleter interface {
	deletePublisher(*Publisher)
	deleteConsumer(*Consumer)
}

type mqChannel interface {
	Close() error
	Consume(string, string, bool, bool, bool, bool, amqp.Table) (<-chan amqp.Delivery, error)
	NotifyClose(chan *amqp.Error) chan *amqp.Error
	Publish(string, string, bool, bool, amqp.Publishing) error
	Qos(int, int, bool) error
}

type mqDeleterTest struct {
	_deletePublisher func(*Publisher)
	_deleteConsumer  func(*Consumer)
}

func (m *mqDeleterTest) deletePublisher(p *Publisher) {
	m._deletePublisher(p)
}

func (m *mqDeleterTest) deleteConsumer(c *Consumer) {
	m._deleteConsumer(c)
}

type mqChannelTest struct {
	_Close       func() error
	_Consume     func(string, string, bool, bool, bool, bool, amqp.Table) (<-chan amqp.Delivery, error)
	_NotifyClose func(chan *amqp.Error) chan *amqp.Error
	_Publish     func(string, string, bool, bool, amqp.Publishing) error
	_Qos         func(int, int, bool) error
}

func (m *mqChannelTest) Close() error {
	return m._Close()
}

func (m *mqChannelTest) Consume(name string, tag string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return m._Consume(name, tag, autoAck, exclusive, noLocal, noWait, args)
}

func (m *mqChannelTest) NotifyClose(c chan *amqp.Error) chan *amqp.Error {
	return m._NotifyClose(c)
}

func (m *mqChannelTest) Publish(exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	return m._Publish(exchange, key, mandatory, immediate, msg)
}

func (m *mqChannelTest) Qos(prefetchCount int, prefetchSize int, global bool) error {
	return m._Qos(prefetchCount, prefetchSize, global)
}
