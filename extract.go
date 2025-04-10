package matroska

import (
	"io"
	"time"
)

type block interface {
	Timestamp(scale time.Duration) time.Duration
	Data() io.Reader
}
