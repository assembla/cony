# Cony

High-level AMQP 0.9.1 client library. It's wrapper around low-level [rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go/) library.
This is a fork from github.com/assembla/cony to change from deprecated [streadway/amqp](https://github.com/streadway/amqp/) library.

# Goals

Provide a way to work with AMQP declaratively

# Requirments

The library uses [atomic.Value](http://golang.org/pkg/sync/atomic/#Value), so Go 1.4+ is needed.

# Documentation

[![GoDoc](https://godoc.org/github.com/assembla/cony?status.svg)](https://godoc.org/github.com/assembla/cony)
[![Build Status](https://travis-ci.org/assembla/cony.svg)](https://travis-ci.org/assembla/cony)

# Thread-safety

Cony is thread-safe as long as [streadway/amqp](https://github.com/streadway/amqp) is thread-safe. It's recommended to open AMQP channel per thread, so in case of `cony` it should be `Consumer` `Producer` per goroutine.

# License

BSD 2 clause - see LICENSE for more details.
