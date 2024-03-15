// Connmanager helps in managing the connections, but allowing you to set up a proxy(iris)
// to connect to a remote server. The connection to the remote server stays as long as the
// client is connected to the proxy(iris).

package connmanager

import (
	"context"
	"sync"
)

var (
	_ ConnCountManager = &SimpleConnCountManager{}
)

type SimpleConnCountManager struct {
	connLimit     int
	connCount     int
	connCountLock sync.RWMutex
	cond          sync.Cond
}

// NewSimpleConnCountManager takes the connLimt as an argument and returns a new instance of
// SimpleConnCountManager struct which represents connection limit, number of connections, locked connection
// count and condition.
func NewSimpleConnCountManager(connLimit int) *SimpleConnCountManager {
	res := &SimpleConnCountManager{
		connCount:     0,
		connLimit:     connLimit,
		connCountLock: sync.RWMutex{},
		cond:          *sync.NewCond(&sync.Mutex{}),
	}
	// sync.Cond must be Locked before Waiting on it
	res.cond.L.Lock()
	return res
}

// getConnCount returns the currnet connection count.
func (c *SimpleConnCountManager) getConnCount() int {
	c.connCountLock.RLock()
	defer c.connCountLock.RUnlock()
	return c.connCount
}

// inc increments the connection count.
func (c *SimpleConnCountManager) inc() {
	c.connCountLock.Lock()
	defer c.connCountLock.Unlock()
	c.connCount++
}

// Acquire returns the status of the connection
// TODO : check the functionality of sync pakage and update comment
func (c *SimpleConnCountManager) Acquire(ctx context.Context) bool {
	if c.getConnCount() < c.connLimit {
		c.inc()
		return true
	}
	var condChan chan bool
	go func() {
		c.cond.Wait()
		condChan <- true
	}()
	select {
	case <-condChan:
		c.inc()
		return true
	case <-ctx.Done():
		return false
	}
}

// Remove removes the latest connection
func (c *SimpleConnCountManager) Remove() {
	c.connCountLock.Lock()
	c.connCount--
	c.connCountLock.Unlock()
	c.cond.Signal()
}
