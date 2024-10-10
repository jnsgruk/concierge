package apt

import (
	"fmt"
	"log/slog"

	"github.com/jnsgruk/concierge/internal/runner"
)

// Update is a helper method to update the host's package cache.
func Update() error {
	_, err := runner.NewCommandSudo("apt", []string{"update"}).Run()
	if err != nil {
		return fmt.Errorf("failed to update apt package lists: %w", err)
	}

	return nil
}

// NewAptPackage constructs a new AptPackage instance.
func NewAptPackage(name string) *AptPackage {
	return &AptPackage{Name: name}
}

// AptPackage is a simple representation of a package installed from the Ubuntu archive.
type AptPackage struct {
	Name string
}

// Install uses `apt` to install the package on the system from the archives.
func (p *AptPackage) Install() error {
	_, err := runner.NewCommandSudo("apt", []string{"install", "-y", p.Name}).Run()
	if err != nil {
		return fmt.Errorf("failed to install apt package '%s': %w", p.Name, err)
	}

	slog.Info("Installed apt package", "package", p.Name)
	return nil
}

// Remove uninstalls the package from the system.
func (p *AptPackage) Remove() error {
	_, err := runner.NewCommandSudo("apt", []string{"remove", "-y", p.Name}).Run()
	if err != nil {
		return fmt.Errorf("failed to remove apt package '%s': %w", p.Name, err)
	}

	return nil
}
