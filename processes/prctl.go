package processes

import (
	"fmt"
	"syscall"
)

func PrctlSetPdeathsig(sig syscall.Signal) error {
	if _, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, syscall.PR_SET_PDEATHSIG, uintptr(sig), 0, 0, 0, 0); errno != 0 {
		return fmt.Errorf("prctl PR_SET_PDEATHSIG %d: %w", sig, syscall.Errno(errno))
	}
	return nil
}
