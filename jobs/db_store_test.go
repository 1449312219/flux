package jobs

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"
	"time"

	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/db"
)

var (
	databaseSource = flag.String("database-source", "", `Database source name. The default is a temporary DB using ql`)

	done        chan error
	errRollback = fmt.Errorf("Rolling back test data")
)

func mkDBFile(t *testing.T) string {
	f, err := ioutil.TempFile("", "fluxy-testdb")
	if err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func bailIfErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

// Setup sets up stuff for testing, creating a new database
func Setup(t *testing.T) *DatabaseStore {
	if *databaseSource == "" {
		*databaseSource = "file://" + mkDBFile(t)
	}

	u, err := url.Parse(*databaseSource)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = db.Migrate(*databaseSource, "../db/migrations"); err != nil {
		t.Fatal(err)
	}

	db, err := NewDatabaseStore(db.DriverForScheme(u.Scheme), *databaseSource, 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	newDB := make(chan *DatabaseStore)
	done = make(chan error)
	go func() {
		done <- db.Transaction(func(tx *DatabaseStore) error {
			// Pass out the tx so we can run the test
			newDB <- tx
			// Wait for the test to finish
			return <-done
		})
	}()
	// Get the new database
	return <-newDB
}

// Cleanup cleans up after a test
func Cleanup(t *testing.T, database *DatabaseStore) {
	if done != nil {
		done <- errRollback
		err := <-done
		if err != errRollback {
			t.Fatalf("Unexpected error %q", err)
		}
		done = nil
	}
}

func TestDatabaseStore(t *testing.T) {
	instance := flux.InstanceID("instance")
	db := Setup(t)
	defer Cleanup(t, db)

	// Get a job when there are none
	_, err := db.NextJob(nil)
	if err != flux.ErrNoJobAvailable {
		t.Fatalf("Expected ErrNoJobAvailable, got %q", err)
	}

	// Put some jobs
	backgroundJobID, err := db.PutJob(instance, flux.Job{
		Method:   flux.ReleaseJob,
		Params:   flux.ReleaseJobParams{},
		Priority: flux.PriorityBackground,
	})
	bailIfErr(t, err)
	interactiveJobID, err := db.PutJob(instance, flux.Job{
		Method:   flux.ReleaseJob,
		Params:   flux.ReleaseJobParams{},
		Priority: flux.PriorityInteractive,
	})
	bailIfErr(t, err)

	// Take one
	interactiveJob, err := db.NextJob(nil)
	bailIfErr(t, err)
	// - It should be the highest priority
	if interactiveJob.ID != interactiveJobID {
		t.Errorf("Got a lower priority job when a higher one was available")
	}
	// - It should have a default queue
	if interactiveJob.Queue != flux.DefaultQueue {
		t.Errorf("job default queue (%q) was not expected (%q)", interactiveJob.Queue, flux.DefaultQueue)
	}
	// - It should have been scheduled in the past
	if interactiveJob.ScheduledAt.IsZero() || interactiveJob.ScheduledAt.After(time.Now()) {
		t.Errorf("expected job to be scheduled in the past")
	}
	// - It should have a log and status
	if len(interactiveJob.Log) == 0 || interactiveJob.Status == "" {
		t.Errorf("expected job to have a log and status")
	}

	// Update the job
	newStatus := "Being used in testing"
	interactiveJob.Status = newStatus
	interactiveJob.Log = append(interactiveJob.Log, newStatus)
	bailIfErr(t, db.UpdateJob(interactiveJob))
	// - It should have saved the changes
	interactiveJob, err = db.GetJob(instance, interactiveJobID)
	bailIfErr(t, err)
	if interactiveJob.Status != newStatus || len(interactiveJob.Log) != 2 || interactiveJob.Log[1] != interactiveJob.Status {
		t.Errorf("expected job to have new log and status")
	}

	// Heartbeat the job
	oldHeartbeat := interactiveJob.Heartbeat
	bailIfErr(t, db.Heartbeat(interactiveJobID))
	// - Heartbeat time should be updated
	interactiveJob, err = db.GetJob(instance, interactiveJobID)
	bailIfErr(t, err)
	if !interactiveJob.Heartbeat.After(oldHeartbeat) {
		t.Errorf("expected job heartbeat to have been updated")
	}

	// Take the next
	backgroundJob, err := db.NextJob(nil)
	bailIfErr(t, err)
	// - It should be different
	if backgroundJob.ID != backgroundJobID {
		t.Errorf("Got a different job than expected")
	}

	// Finish one
	backgroundJob.Done = true
	backgroundJob.Success = true
	bailIfErr(t, db.UpdateJob(backgroundJob))
	// - Status should be changed
	backgroundJob, err = db.GetJob(instance, backgroundJobID)
	bailIfErr(t, err)
	if !backgroundJob.Done || !backgroundJob.Success {
		t.Errorf("expected job to have been marked as done")
	}

	// GC
	// - Advance time so we can gc stuff
	db.now = func(_ dbProxy) (time.Time, error) {
		return time.Now().Add(2 * time.Minute), nil
	}
	bailIfErr(t, db.GC())
	// - Finished should be removed
	_, err = db.GetJob(instance, backgroundJobID)
	if err != flux.ErrNoSuchJob {
		t.Errorf("expected ErrNoSuchJob, got %q", err)
	}
}

func TestDatabaseStoreScheduledJobs(t *testing.T) {
	instance := flux.InstanceID("instance")
	db := Setup(t)
	defer Cleanup(t, db)

	// Put a scheduled job
	scheduledJobID, err := db.PutJob(instance, flux.Job{
		Method:      flux.ReleaseJob,
		Params:      flux.ReleaseJobParams{},
		ScheduledAt: time.Now().Add(1 * time.Minute),
	})

	// Check it isn't available
	if _, err := db.NextJob(nil); err != flux.ErrNoJobAvailable {
		t.Fatalf("Expected ErrNoJobAvailable, got %q", err)
	}

	// Advance time so it it available
	db.now = func(_ dbProxy) (time.Time, error) {
		return time.Now().Add(2 * time.Minute), nil
	}

	// It should be available
	job, err := db.NextJob(nil)
	bailIfErr(t, err)
	if job.ID != scheduledJobID {
		t.Fatalf("Expected scheduled job, got %q", job.ID)
	}
}
