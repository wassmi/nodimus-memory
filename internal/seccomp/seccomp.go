//go:build linux

package seccomp

import (
	"fmt"

	"github.com/seccomp/libseccomp-golang"
)

// LoadProfile loads the seccomp profile.
func LoadProfile() error {
	filter, err := seccomp.NewFilter(seccomp.ActAllow)
	if err != nil {
		return fmt.Errorf("failed to create seccomp filter: %w", err)
	}

	// Add rules to the filter here.
	// For example, to deny the "execve" syscall:
	// execveSyscall, err := seccomp.GetSyscallFromName("execve")
	// if err != nil {
	// 	return fmt.Errorf("failed to get syscall number: %w", err)
	// }
	// if err := filter.AddRule(execveSyscall, seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))); err != nil {
	// 	return fmt.Errorf("failed to add seccomp rule: %w", err)
	// }

	if err := filter.Load(); err != nil {
		return fmt.Errorf("failed to load seccomp filter: %w", err)
	}

	return nil
}