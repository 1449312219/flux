package release

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/weaveworks/flux/resource"
	"github.com/weaveworks/flux/update"
)

type Changes interface {
	CalculateRelease(update.ReleaseContext, log.Logger) ([]*update.ControllerUpdate, update.Result, error)
	ReleaseKind() update.ReleaseKind
	ReleaseType() update.ReleaseType
	CommitMessage(update.Result) string
}

func Release(rc *ReleaseContext, changes Changes, logger log.Logger) (results update.Result, err error) {
	defer func(start time.Time) {
		update.ObserveRelease(
			start,
			err == nil,
			changes.ReleaseType(),
			changes.ReleaseKind(),
		)
	}(time.Now())

	logger = log.With(logger, "type", "release")

	before, err := rc.manifests.LoadManifests(rc.repo.Dir(), rc.repo.ManifestDir())
	updates, results, err := changes.CalculateRelease(rc, logger)
	if err != nil {
		return nil, err
	}

	err = ApplyChanges(rc, updates, logger)
	if err == nil {
		var after map[string]resource.Resource
		after, err = rc.manifests.LoadManifests(rc.repo.Dir(), rc.repo.ManifestDir())
		if err == nil {
			err = VerifyChanges(before, updates, after)
		}
	}
	if err != nil {
		err = MakeReleaseError(err)
	}
	return results, err
}

func ApplyChanges(rc *ReleaseContext, updates []*update.ControllerUpdate, logger log.Logger) error {
	logger.Log("updates", len(updates))
	if len(updates) == 0 {
		logger.Log("exit", "no images to update for services given")
		return nil
	}

	timer := update.NewStageTimer("write_changes")
	err := rc.WriteUpdates(updates)
	timer.ObserveDuration()
	return err
}

// VerifyChanges checks that the `after` resources are exactly the
// `before` resources with the updates applied. It destructively
// updates `before`.
func VerifyChanges(before map[string]resource.Resource, updates []*update.ControllerUpdate, after map[string]resource.Resource) error {
	timer := update.NewStageTimer("verify_changes")
	defer timer.ObserveDuration()

	for _, update := range updates {
		res, ok := before[update.ResourceID.String()]
		if !ok {
			return fmt.Errorf("resource %q mentioned in update not found in resources", update.ResourceID.String())
		}
		wl, ok := res.(resource.Workload)
		if !ok {
			return fmt.Errorf("resource %q mentioned in update is not a workload", update.ResourceID.String())
		}
		for _, containerUpdate := range update.Updates {
			wl.SetContainerImage(containerUpdate.Container, containerUpdate.Target)
		}
	}

	for id, afterRes := range after {
		beforeRes, ok := before[id]
		if !ok {
			return fmt.Errorf("resource %q is new after update")
		}
		delete(before, id)

		beforeWorkload, ok := beforeRes.(resource.Workload)
		if !ok {
			// the resource in question isn't a workload, so ignore it.
			continue
		}
		afterWorkload, ok := afterRes.(resource.Workload)
		if !ok {
			return fmt.Errorf("resource %q is no longer a workload (Deployment or DaemonSet, or similar) after update", id)
		}

		beforeContainers := beforeWorkload.Containers()
		afterContainers := afterWorkload.Containers()
		if len(beforeContainers) != len(afterContainers) {
			return fmt.Errorf("resource %q has different set of containers after update", id)
		}
		for i := range afterContainers {
			if beforeContainers[i].Name != afterContainers[i].Name {
				return fmt.Errorf("Container in position %d of resource %q has a different name after update: was %q, now %q", i, id, beforeContainers[i].Name, afterContainers[i].Name)
			}
			if beforeContainers[i].Image != afterContainers[i].Image {
				return fmt.Errorf("The image for container %q in resource %q should be %q, but is %q", beforeContainers[i].Name, id, beforeContainers[i].Image.String(), afterContainers[i].Image.String())
			}
		}

	}

	var disappeared []string
	for id := range before {
		disappeared = append(disappeared, fmt.Sprintf("%q", id))
	}
	if len(disappeared) > 0 {
		return fmt.Errorf("resources {%s} present before update but not after", strings.Join(disappeared, ", "))
	}

	return nil
}
