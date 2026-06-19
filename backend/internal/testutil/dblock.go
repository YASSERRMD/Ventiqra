// Package testutil provides shared helpers for integration tests.
//
// Because Go runs each package's tests in a separate binary, and Ventiqra's
// integration tests share a single PostgreSQL database with per-test schema
// resets, tests across packages can otherwise interleave (one package's
// `DROP SCHEMA` clobbering another's in-flight assertions). The helpers here
// serialize the destructive phases across processes via an exclusive file lock.
package testutil

import (
	"os"
	"path/filepath"
	"syscall"
)

// dbLockFileName is the shared lock file (under the OS temp dir) used to
// serialize database-resetting integration tests across test binaries.
const dbLockFileName = "ventiqra-test-db.lock"

// LockDB acquires an exclusive, process-wide file lock used to serialize
// integration tests that reset and migrate the shared test database. The
// returned release function MUST be called when the test finishes; callers
// typically register it with t.Cleanup. The lock is held for the test's full
// duration so that a schema reset in one package cannot race with assertions
// in another.
//
// If the lock cannot be acquired (for example on an unsupported platform), a
// no-op release is returned so the test still runs, just without serialization.
func LockDB() func() {
	path := filepath.Join(os.TempDir(), dbLockFileName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return func() {}
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return func() {}
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
}
