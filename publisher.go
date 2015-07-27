package cony

import "github.com/streadway/amqp"

// PublisherOpt is a functional option type for Publisher
type PublisherOpt func(*Publisher)

type publishMaybeErr struct {
	pub chan amqp.Publishing
	err chan error
}

// Publisher hold definition for AMQP publishing
type Publisher struct {
	exchange string
	key      string
	tmpl     amqp.Publishing
	pubChan  chan publishMaybeErr
	stop     chan struct{}
}

// Template will be used, input buffer will be added as Publishing.Body.
// return int will always be len(b)
//
// Implements io.Writer
func (p *Publisher) Write(b []byte) (int, error) {
	pub := p.tmpl
	pub.Body = b
	return len(b), p.Publish(pub)
}

// Publish used to publish custom amqp.Publishing
func (p *Publisher) Publish(pub amqp.Publishing) error {
	reqRepl := publishMaybeErr{
		pub: make(chan amqp.Publishing, 2),
		err: make(chan error, 2),
	}

	reqRepl.pub <- pub
	p.pubChan <- reqRepl
	err := <-reqRepl.err
	return err
}

// Cancel this publisher
func (p *Publisher) Cancel() {
	select {
	case p.stop <- struct{}{}:
	default:
	}
}

func (p *Publisher) serve(client *Client) {
	ch, err := client.channel()
	if err != nil {
		return
	}

	chanErrs := make(chan *amqp.Error)
	ch.NotifyClose(chanErrs)

	for {
		select {
		case <-p.stop:
			client.deletePublisher(p)
			ch.Close()
			return
		case <-chanErrs:
			return
		case envelop := <-p.pubChan:
			msg := <-envelop.pub
			close(envelop.pub)
			if err := ch.Publish(p.exchange, p.key, false, false, msg); err != nil {
				envelop.err <- err
			}
			close(envelop.err)
		}
	}
}

// NewPublisher is a Publisher constructor
func NewPublisher(exchange string, key string, opts ...PublisherOpt) *Publisher {
	p := &Publisher{
		exchange: exchange,
		key:      key,
		pubChan:  make(chan publishMaybeErr),
		stop:     make(chan struct{}, 2),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

// PublishingTemplate Publisher's functional option. Provide template
// amqp.Publishing and save typing.
func PublishingTemplate(t amqp.Publishing) PublisherOpt {
	return func(p *Publisher) {
		p.tmpl = t
	}
}
