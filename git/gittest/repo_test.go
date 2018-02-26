package gittest

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"context"

	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/cluster/kubernetes/testfiles"
	"github.com/weaveworks/flux/git"
	"github.com/weaveworks/flux/job"
	"github.com/weaveworks/flux/update"
)

func TestCheckout(t *testing.T) {
	repo, cleanup := Repo(t)
	defer cleanup()

	sd, sg := make(chan struct{}), &sync.WaitGroup{}

	sg.Add(1)
	go repo.Start(sd, sg)
	WaitForRepoReady(repo, t)

	ctx := context.Background()

	params := git.Config{
		Branch:    "master",
		UserName:  "example",
		UserEmail: "example@example.com",
		SyncTag:   "flux-test",
		NotesRef:  "fluxtest",
	}
	checkout, err := repo.Clone(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	defer checkout.Clean()

	// We don't expect any notes in the clone, yet. Make sure we get
	// no note, rather than an error.
	head, err := checkout.HeadRevision(ctx)
	if err != nil {
		t.Fatal(err)
	}
	note, err := checkout.GetNote(ctx, head)
	if err != nil {
		t.Error(err)
	}
	if note != nil {
		t.Errorf("Expected no note on head revision; got %#v", note)
	}

	changedFile := ""
	for file, _ := range testfiles.Files {
		path := filepath.Join(checkout.ManifestDir(), file)
		if err := ioutil.WriteFile(path, []byte("FIRST CHANGE"), 0666); err != nil {
			t.Fatal(err)
		}
		changedFile = file
		break
	}
	commitAction := &git.CommitAction{Author: "", Message: "Changed file"}
	if err := checkout.CommitAndPush(ctx, commitAction, nil); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(checkout.ManifestDir(), changedFile)
	if err := ioutil.WriteFile(path, []byte("SECOND CHANGE"), 0666); err != nil {
		t.Fatal(err)
	}
	// An example note with some of the fields filled in, so we can test
	// serialization a bit.
	expectedNote := git.Note{
		JobID: job.ID("jobID1234"),
		Spec: update.Spec{
			Type: update.Images,
			Spec: update.ReleaseSpec{},
		},
		Result: update.Result{
			flux.MustParseResourceID("default/service1"): update.ControllerResult{
				Status: update.ReleaseStatusFailed,
				Error:  "failed the frobulator",
			},
		},
	}
	commitAction = &git.CommitAction{Author: "", Message: "Changed file again"}
	if err := checkout.CommitAndPush(ctx, commitAction, &expectedNote); err != nil {
		t.Fatal(err)
	}

	check := func(c *git.Checkout) {
		contents, err := ioutil.ReadFile(filepath.Join(c.ManifestDir(), changedFile))
		if err != nil {
			t.Fatal(err)
		}
		if string(contents) != "SECOND CHANGE" {
			t.Error("contents in checkout are not what we committed")
		}
		rev, err := c.HeadRevision(ctx)
		if err != nil {
			t.Fatal(err)
		}
		note, err := c.GetNote(ctx, rev)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(*note, expectedNote) {
			t.Errorf("note is not what we supplied when committing: %#v", note)
		}
	}

	// Do we see the changes if we make another working checkout?
	if err := repo.Refresh(ctx); err != nil {
		t.Error(err)
	}

	another, err := repo.Clone(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	defer another.Clean()
	check(another)

	close(sd)
	sg.Wait()
}
