package instance

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/pkg/errors"

	"github.com/weaveworks/fluxy"
	"github.com/weaveworks/fluxy/git"
	"github.com/weaveworks/fluxy/history"
	"github.com/weaveworks/fluxy/platform"
	"github.com/weaveworks/fluxy/registry"
)

type MultitenantInstancer struct {
	DB        DB
	Connecter platform.Connecter
	Logger    log.Logger
	Histogram metrics.Histogram
	History   history.DB
}

func (m *MultitenantInstancer) Get(instanceID flux.InstanceID) (*Instance, error) {
	c, err := m.DB.GetConfig(instanceID)
	if err != nil {
		return nil, errors.Wrap(err, "getting instance config from DB")
	}

	// Platform interface for this instance
	platform, err := m.Connecter.Open(instanceID)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to platform")
	}

	// Logger specialised to this instance
	instanceLogger := log.NewContext(m.Logger).With("instanceID", instanceID)

	// Registry client with instance's config
	creds := credentialsFromConfig(c.Settings)
	regClient := &registry.Client{
		Credentials: creds,
		Logger:      instanceLogger,
	}

	repo := gitRepoFromSettings(c.Settings)

	// Events for this instance
	events := EventReadWriter{instanceID, m.History}
	// Configuration for this instance
	config := InstanceConfig{instanceID, m.DB}

	return New(
		platform,
		regClient,
		config,
		repo,
		instanceLogger,
		m.Histogram,
		events,
		events,
	), nil
}

func credentialsFromConfig(config flux.InstanceConfig) registry.Credentials {
	return registry.NoCredentials() // %%% FIXME
}

func gitRepoFromSettings(settings flux.InstanceConfig) git.Repo {
	branch := settings.Git.Branch
	if branch == "" {
		branch = "master"
	}
	return git.Repo{
		URL:    settings.Git.URL,
		Branch: branch,
		Key:    settings.Git.Key,
		Path:   settings.Git.Path,
	}
}
