package kubernetes

import (
	"github.com/weaveworks/flux"
	kresource "github.com/weaveworks/flux/cluster/kubernetes/resource"
	"github.com/weaveworks/flux/image"
	"github.com/weaveworks/flux/resource"
)

type Manifests struct {
	FallbackNamespace string
}

func (c *Manifests) LoadManifests(base string, paths []string) (map[string]resource.Resource, error) {
	return kresource.Load(base, c.FallbackNamespace, paths)
}

func (c *Manifests) UpdateImage(def []byte, id flux.ResourceID, container string, image image.Ref) ([]byte, error) {
	return updatePodController(def, id, container, image)
}

// UpdatePolicies and ServicesWithPolicies in policies.go
