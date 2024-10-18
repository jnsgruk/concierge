package concierge

import (
	"fmt"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/juju"
	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/runner"
	"golang.org/x/sync/errgroup"
)

// Plan represents a set of packages and providers that are to be prepared/restored.
type Plan struct {
	Providers []providers.Provider
	Snaps     []packages.SnapPackage
	Debs      []*packages.Deb

	config *config.Config
	runner runner.CommandRunner
}

// NewPlan constructs a new plan consisting of snaps/debs/providers & juju.
func NewPlan(config *config.Config, runner runner.CommandRunner) *Plan {
	plan := &Plan{config: config, runner: runner}

	for _, s := range append(config.Host.Snaps, config.Overrides.ExtraSnaps...) {
		snap := packages.NewSnapFromString(s)

		// Check if the channel has been overridden by a CLI argument/env var
		channelOverride := getSnapChannelOverride(config, snap.Name())
		if channelOverride != "" {
			snap.SetChannel(channelOverride)
		}

		plan.Snaps = append(plan.Snaps, snap)
	}

	for _, p := range append(config.Host.Packages, config.Overrides.ExtraDebs...) {
		plan.Debs = append(plan.Debs, packages.NewDeb(p))
	}

	for _, providerName := range providers.SupportedProviders {
		if p := providers.NewProvider(providerName, runner, config); p != nil {
			plan.Providers = append(plan.Providers, p)
		}
	}

	return plan
}

// Execute either prepares or restores a given plan
func (p *Plan) Execute(action string) error {
	var eg errgroup.Group

	snapHandler := packages.NewSnapHandler(p.runner, p.Snaps)
	debHandler := packages.NewDebHandler(p.runner, p.Debs)
	jujuHandler := juju.NewJujuHandler(p.config, p.runner, p.Providers)

	// Prepare/restore package handlers concurrently
	eg.Go(func() error { return DoAction(snapHandler, action) })
	eg.Go(func() error { return DoAction(debHandler, action) })
	if err := eg.Wait(); err != nil {
		return err
	}

	// Prepare/restore providers concurrently
	for _, provider := range p.Providers {
		eg.Go(func() error { return DoAction(provider, action) })
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	// Prepare/Restore juju controllers
	err := DoAction(jujuHandler, action)
	if err != nil {
		return fmt.Errorf("failed to prepare Juju: %w", err)
	}

	return nil
}

// getSnapChannelOverride takes the name of a snap. If the snap's version
// is overridden, the overridden channel is returned.
func getSnapChannelOverride(config *config.Config, snap string) string {
	switch snap {
	case "charmcraft":
		return config.Overrides.CharmcraftChannel
	case "snapcraft":
		return config.Overrides.SnapcraftChannel
	case "rockcraft":
		return config.Overrides.RockcraftChannel
	default:
		return ""
	}
}
