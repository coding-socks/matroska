package matroska

import (
	"github.com/coding-socks/ebml"
	"io"
	"log"
)

type ClusterScanner interface {
	Next() bool
	Cluster() Cluster
	Err() error
}

type clusterScanner struct {
	d      *ebml.Decoder
	el     ebml.Element
	offset int64

	cluster Cluster
	err     error
}

func NewClusterScanner(d *ebml.Decoder, el ebml.Element, offset int64) ClusterScanner {
	return &clusterScanner{d: d, el: el, offset: offset}
}

func (c *clusterScanner) Next() bool {
	d := c.d
	segmentEl := c.el
	offset := c.offset
	for {
		if ok, _ := d.EndOfElement(segmentEl, offset); ok {
			// TODO: check return val
			return false
		}
		el, n, err := d.Next()
		if segmentEl.DataSize.Known() {
			offset += int64(n)
		}
		if err == ebml.ErrInvalidVINTLength {
			continue
		} else if err != nil {
			c.err = err
			return false
		}
		if segmentEl.DataSize.Known() {
			offset += el.DataSize.Size()
		}
		switch el.ID {
		default:
			if _, err := d.Seek(el.DataSize.Size(), io.SeekCurrent); err != nil {
				log.Fatalf("Could not skip %s: %s", el.ID, err)
			}
		case IDCluster:
			var cl Cluster
			if err := d.Decode(&cl); err != nil {
				log.Fatalf("Could not decode %s: %s", el.ID, err)
			}
			c.cluster = cl
			return true
		}
	}
}

func (c *clusterScanner) Cluster() Cluster {
	return c.cluster
}

func (c *clusterScanner) Err() error {
	return c.err
}
