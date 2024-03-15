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

func NewSimpleConnCountManager(connLimit int) *SimpleConnCountManager {
	return &SimpleConnCountManager{
		connCount:     0,
		connLimit:     connLimit,
		connCountLock: sync.RWMutex{},
		cond:          *sync.NewCond(&sync.Mutex{}),
	}
}

func (c *SimpleConnCountManager) getConnCount() int {
	c.connCountLock.RLock()
	defer c.connCountLock.RUnlock()
	return c.connCount
}

func (c *SimpleConnCountManager) inc() {
	c.connCountLock.Lock()
	defer c.connCountLock.Unlock()
	c.connCount++
}

func (c *SimpleConnCountManager) Acquire(ctx context.Context) bool {
	if c.getConnCount() < c.connLimit {
		c.inc()
		return true
	}
	var condChan chan bool
	go func() {
		c.cond.L.Lock()
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

func (c *SimpleConnCountManager) Remove() {
	c.connCountLock.Lock()
	c.connCount--
	c.connCountLock.Unlock()
	c.cond.Signal()
}
