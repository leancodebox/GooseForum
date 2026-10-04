package mfaservice

import "sync"

// Bounded stripes serialize a user's MFA mutations without retaining user IDs.
// Like the current sharedcache backend, these locks are process-local.
var operationLocks [256]sync.Mutex

func lockOperation(userID uint64) func() {
	lock := &operationLocks[userID%uint64(len(operationLocks))]
	lock.Lock()
	return lock.Unlock
}
