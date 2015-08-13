# Examples

## Basic

This is the base example ment to show how to use cony.
The producer will exit after each publish action finishes.
The consumer will run forever (or until you stop the program with CTRL-C).
Both applications are CLI-based.

[View producer code](basic/producer/producer.go)
[View consumer code](basic/consumer/consumer.go)

## Web

Example that illustrates how to use cony with a web-flow.
The producer is a simple web application. It allows to set the body of a "publishing" from a web-form.
The consumer is a CLI-application that shows the contents of each publishing it receives.

[View producer code](web/producer/producer.go)
[View consumer code](web/consumer/consumer.go)

## Email

Example that illustrates having multiple workers handling a particular job.
In this case the producer is used to show a simple form for filling in some email fields.
The consumer launches a gorouting that receives deliveries and then balances these between 3 worker goroutines that actually do the email sending.

[View producer code](email/producer/producer.go)
[View consumer code](email/consumer/consumer.go)
