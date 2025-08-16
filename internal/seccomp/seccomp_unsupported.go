//go:build !linux

package seccomp

// LoadProfile is a no-op on non-Linux systems.
func LoadProfile() error {
	return nil
}
