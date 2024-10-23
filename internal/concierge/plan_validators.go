package concierge

import (
	"fmt"
	"slices"
)

// planValidators is a list of planValidators used to verify a plan
var planValidators = []func(p *Plan) error{
	validateSingleLocalKubernetesInstance,
}

// validateSingleLocalKubernetesInstance ensures the plan won't try and install multiple
// local Kubernetes providers, which would conflict.
func validateSingleLocalKubernetesInstance(plan *Plan) error {
	providerNames := []string{}

	for _, p := range plan.Providers {
		providerNames = append(providerNames, p.Name())
	}

	if slices.Contains(providerNames, "microk8s") && slices.Contains(providerNames, "canonical-k8s") {
		return fmt.Errorf("cannot configure multiple local kubernetes providers")
	}

	return nil
}
