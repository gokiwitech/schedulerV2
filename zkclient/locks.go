package zkclient

import (
	"path"

	"github.com/samuel/go-zookeeper/zk"
)

type DistributedLock struct {
	Conn         *zk.Conn
	LockBasePath string
	LockName     string
	LockPath     string
	Acl          []zk.ACL
}

func (dl *DistributedLock) Acquire() (bool, error) {
	// Ensure the base path exists
	_, err := dl.Conn.Create(dl.LockBasePath, []byte{}, 0, dl.Acl)
	if err != nil && err != zk.ErrNodeExists {
		return false, err
	}

	// Construct the path for the lock node using the message ID
	dl.LockPath = path.Join(dl.LockBasePath, dl.LockName)

	// Attempt to create an ephemeral node for the lock
	_, err = dl.Conn.Create(dl.LockPath, []byte{}, zk.FlagEphemeral, dl.Acl)
	if err == zk.ErrNodeExists {
		// The node already exists, which means the lock is already held by another process
		return false, nil
	} else if err != nil {
		// Some other error occurred while trying to create the node
		return false, err
	}

	// The lock was successfully acquired
	return true, nil
}

func (dl *DistributedLock) Release() error {
	if dl.LockPath == "" {
		// Lock was never acquired or already released
		return nil
	}
	err := dl.Conn.Delete(dl.LockPath, -1)
	dl.LockPath = "" // Clear the lock path after releasing
	return err
}
