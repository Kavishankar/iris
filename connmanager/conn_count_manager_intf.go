package connmanager

import "context"

type ConnCountManager interface {
	Acquire(context.Context) bool
	Remove()
}
