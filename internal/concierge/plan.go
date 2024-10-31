package concierge

import (
	"fmt"

	"github.com/jnsgruk/concierge/internal/config"
	"github.com/jnsgruk/concierge/internal/juju"
	"github.com/jnsgruk/concierge/internal/packages"
	"github.com/jnsgruk/concierge/internal/providers"
	"github.com/jnsgruk/concierge/internal/system"
	"golang.org/x/sync/errgroup"
)

// Plan represents a set of packages and providers that are to be prepared/restored.
type Plan struct {
	Providers []providers.Provider
	Snaps     []*system.Snap
	Debs      []*packages.Deb

	config *config.Config
	system system.Worker
}

// NewPlan constructs a new plan consisting of snaps/debs/providers & juju.
func NewPlan(config *config.Config, system system.Worker) *Plan {
	plan := &Plan{config: config, system: system}

	for _, s := range append(config.Host.Snaps, config.Overrides.ExtraSnaps...) {
		snap := system.NewSnapFromString(s)

		// Check if the channel has been overridden by a CLI argument/env var
		channelOverride := getSnapChannelOverride(config, snap.Name)
		if channelOverride != "" {
			snap.Channel = channelOverride
		}

		plan.Snaps = append(plan.Snaps, snap)
	}

	for _, p := range append(config.Host.Packages, config.Overrides.ExtraDebs...) {
		plan.Debs = append(plan.Debs, packages.NewDeb(p))
	}

	for _, providerName := range providers.SupportedProviders {
		if p := providers.NewProvider(providerName, system, config); p != nil {
			plan.Providers = append(plan.Providers, p)
		}
	}

	return plan
}

// Execute either prepares or restores a given plan
func (p *Plan) Execute(action string) error {
	err := p.validate()
	if err != nil {
		return fmt.Errorf("failed to validate plan: %w", err)
	}

	var eg errgroup.Group

	snapHandler := packages.NewSnapHandler(p.system, p.Snaps)
	debHandler := packages.NewDebHandler(p.system, p.Debs)
	jujuHandler := juju.NewJujuHandler(p.config, p.system, p.Providers)

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
	err = DoAction(jujuHandler, action)
	if err != nil {
		return fmt.Errorf("failed to prepare Juju: %w", err)
	}

	return nil
}

// validate returns an error if the generated plan contains errors that would prevent a successful
// configuration of the machine.
func (p *Plan) validate() error {
	var eg errgroup.Group

	// Run the validators in parallel in an errgroup
	for _, v := range planValidators {
		eg.Go(func() error { return v(p) })
	}
	if err := eg.Wait(); err != nil {
		return err
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
