package v1alpha1

import (
	core "k8s.io/api/core/v1"
)

type ImageRef struct {
	Repository string          `json:"repository,omitempty"`
	Tag        string          `json:"tag,omitempty"`
	PullPolicy core.PullPolicy `json:"pullPolicy,omitempty"`
}
