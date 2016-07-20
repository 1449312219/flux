package main

import (
	"github.com/spf13/cobra"
)

type repoOpts struct {
	*rootOpts
	repository string
}

func newRepo(parent *rootOpts) *repoOpts {
	return &repoOpts{rootOpts: parent}
}

func (opts *repoOpts) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Subcommands dealing with image repositories, e.g., quay.io/weaveworks/helloworld",
	}
	cmd.PersistentFlags().StringVar(&opts.repository, "repo", "", "The repository in question, e.g., quay.io/weaveworks/helloworld")
	return cmd
}
