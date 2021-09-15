package matroska

type ClusterScanner interface {
	Next() bool
	Cluster() Cluster
	Err() error
}

type ClusterSliceScanner struct {
	i  int
	cs []Cluster
}

func NewClusterSliceScanner(cs []Cluster) *ClusterSliceScanner {
	return &ClusterSliceScanner{cs: cs}
}

func (c *ClusterSliceScanner) Next() bool {
	if c.i >= (len(c.cs) - 1) {
		return false
	}
	c.i++
	return true
}

func (c *ClusterSliceScanner) Cluster() Cluster {
	return c.cs[c.i]
}

func (c *ClusterSliceScanner) Err() error {
	return nil
}

type ClusterChannelScanner struct {
	ch   <-chan Cluster
	next Cluster
	err  error
}

func NewClusterChannelScanner(ch <-chan Cluster) *ClusterChannelScanner {
	return &ClusterChannelScanner{ch: ch}
}

func (cs *ClusterChannelScanner) Next() bool {
	c, ok := <-cs.ch
	cs.next = c
	return ok
}

func (cs *ClusterChannelScanner) Cluster() Cluster {
	return cs.next
}

func (cs *ClusterChannelScanner) Err() error {
	return cs.err
}

func (cs *ClusterChannelScanner) Go(f func() error) {
	go func() {
		cs.err = f()
	}()
}
