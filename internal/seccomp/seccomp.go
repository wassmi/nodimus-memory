//go:build linux

package seccomp

import (
	"fmt"

	"github.com/seccomp/libseccomp-golang"
)

// LoadProfile loads the seccomp profile.
func LoadProfile() error {
	filter, err := libseccomp.NewFilter(libseccomp.ActAllow)
	if err != nil {
		return fmt.Errorf("failed to create seccomp filter: %w", err)
	}

	// Add rules to the filter here.
	// For example, to deny the "execve" syscall:
	// if err := filter.AddRule(libseccomp.ScmpSyscall("execve"), libseccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))); err != nil {
	// 	return fmt.Errorf("failed to add seccomp rule: %w", err)
	// }

	if err := filter.Load(); err != nil {
		return fmt.Errorf("failed to load seccomp filter: %w", err)
	}

	return nil
}