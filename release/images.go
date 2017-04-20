package release

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/platform"
	"github.com/weaveworks/flux/registry"
)

// Get the images available for the services given. An image may be
// mentioned more than once in the services, but will only be fetched
// once.
func CollectAvailableImages(reg registry.Registry, services []platform.Service) (ImageMap, error) {
	images := ImageMap{}
	for _, service := range services {
		for _, container := range service.ContainersOrNil() {
			id, err := flux.ParseImageID(container.Image)
			if err != nil {
				// container is running an invalid image id? what?
				return nil, err
			}
			images[id.Repository()] = nil
		}
	}
	for repo := range images {
		r, err := registry.ParseRepository(repo)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing repository %s", repo)
		}
		imageRepo, err := reg.GetRepository(r)
		if err != nil {
			return nil, errors.Wrapf(err, "fetching image metadata for %s", repo)
		}
		res := make([]flux.ImageDescription, len(imageRepo))
		for i, im := range imageRepo {
			id, err := flux.ParseImageID(im.String())
			if err != nil {
				// registry returned an invalid image id
				return nil, err
			}
			res[i] = flux.ImageDescription{
				ID:        id,
				CreatedAt: im.CreatedAt,
			}
		}
		images[repo] = res
	}
	return images, nil
}

// LatestImage returns the latest releasable image for a repository.
// A releasable image is one that is not tagged "latest". (Assumes the
// available images are in descending order of latestness.) If no such
// image exists, returns nil, and the caller can decide whether that's
// an error or not.
func (m ImageMap) LatestImage(repo string) *flux.ImageDescription {
	for _, image := range m[repo] {
		_, _, tag := image.ID.Components()
		if strings.EqualFold(tag, "latest") {
			continue
		}
		return &image
	}
	return nil
}

// For keeping track of which images are available
type ImageMap map[string][]flux.ImageDescription

// Create a map of images. It will check that each image exists.
func ExactImages(reg registry.Registry, images []flux.ImageID) (ImageMap, error) {
	m := ImageMap{}
	for _, id := range images {
		// We must check that the exact images requested actually exist. Otherwise we risk pushing invalid images to git.
		exist, err := imageExists(reg, id)
		if err != nil {
			return m, errors.Wrap(flux.ErrInvalidImageID, err.Error())
		}
		if !exist {
			return m, errors.Wrap(flux.ErrInvalidImageID, fmt.Sprintf("image %q does not exist", id))
		}
		m[id.Repository()] = []flux.ImageDescription{flux.ImageDescription{ID: id}}
	}
	return m, nil
}

// Checks whether the given image exists in the repository.
// Return true if exist, false otherwise
func imageExists(reg registry.Registry, imageID flux.ImageID) (bool, error) {
	// Use this method to parse the image, because it is safe. I.e. it will error and inform the user if it is malformed.
	img, err := flux.ParseImage(imageID.String(), nil)
	if err != nil {
		return false, err
	}
	// Get a specific image.
	_, err = reg.GetImage(registry.RepositoryFromImage(img), img.Tag)
	if err != nil {
		return false, nil
	}
	return true, nil
}
