package cony

import (
	"io"
	"testing"
)

func TestPublisherImplements_io_Writer(t *testing.T) {
	var _ io.Writer = &Publisher{}
}

func TestWrite(t *testing.T)              {}
func TestPublish(t *testing.T)            {}
func TestCancel(t *testing.T)             {}
func TestNewPublisher(t *testing.T)       {}
func TestPublishingTemplate(t *testing.T) {}
