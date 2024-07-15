package common

import (
	"fmt"
	"github.com/harbi-network/harbid/domain/dagconfig"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
)

// RunharbidForTesting runs harbid for testing purposes
func RunharbidForTesting(t *testing.T, testName string, rpcAddress string) func() {
	appDir, err := TempDir(testName)
	if err != nil {
		t.Fatalf("TempDir: %s", err)
	}

	harbidRunCommand, err := StartCmd("harbid",
		"harbid",
		NetworkCliArgumentFromNetParams(&dagconfig.DevnetParams),
		"--appdir", appDir,
		"--rpclisten", rpcAddress,
		"--loglevel", "debug",
	)
	if err != nil {
		t.Fatalf("StartCmd: %s", err)
	}
	t.Logf("harbid started with --appdir=%s", appDir)

	isShutdown := uint64(0)
	go func() {
		err := harbidRunCommand.Wait()
		if err != nil {
			if atomic.LoadUint64(&isShutdown) == 0 {
				panic(fmt.Sprintf("harbid closed unexpectedly: %s. See logs at: %s", err, appDir))
			}
		}
	}()

	return func() {
		err := harbidRunCommand.Process.Signal(syscall.SIGTERM)
		if err != nil {
			t.Fatalf("Signal: %s", err)
		}
		err = os.RemoveAll(appDir)
		if err != nil {
			t.Fatalf("RemoveAll: %s", err)
		}
		atomic.StoreUint64(&isShutdown, 1)
		t.Logf("harbid stopped")
	}
}
