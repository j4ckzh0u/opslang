//go:build !windows

package security

import (
	"os"
	"os/signal"
	"syscall"
)

func ensureGlobalSignalHandler() {
	globalSigOnce.Do(func() {
		signal.Notify(globalSigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-globalSigChan
			globalTempDirs.Range(func(key, value any) bool {
				if td, ok := value.(*TempDir); ok {
					td.Cleanup()
				}
				return true
			})
			signal.Reset(syscall.SIGINT, syscall.SIGTERM)
			_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
		}()
	})
}
