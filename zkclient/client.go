package zkclient

import (
	"github.com/samuel/go-zookeeper/zk"
)

const (
	LockBasePath = "/message_queue_lock"
	LockName     = "lock"
)

// NewDistributedLock creates a new instance of DistributedLock
func NewDistributedLock(conn *zk.Conn, lockBasePath, lockName string) *DistributedLock {
	return &DistributedLock{
		Conn:         conn,
		LockBasePath: lockBasePath,
		LockName:     lockName,
		Acl:          zk.WorldACL(zk.PermAll),
	}
}
