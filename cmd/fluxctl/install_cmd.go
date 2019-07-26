package main

import (
	"github.com/spf13/cobra"
)

type installOpts struct {
	gitURL             string
	gitBranch          string
	gitPaths           []string
	gitLabel           string
	gitUser            string
	gitEmail           string
	namespace          string
	additionalFluxArgs []string
}

func newInstall() *installOpts {
	return &installOpts{}
}

func (opts *installOpts) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Print and tweak Kubernetes manifests needed to install Flux in a Cluster",
		Example: `# Install Flux and make it use Git repository git@github.com:<your username>/flux-get-started
fluxctl install --git-url 'git@github.com:<your username>/flux-get-started' | kubectl -f -`,
		RunE: opts.RunE,
	}
	cmd.Flags().StringVarP(&opts.gitURL, "git-url", "", "",
		"URL of the Git repository to be used by Flux, e.g. git@github.com:<your username>/flux-get-started")
	cmd.Flags().StringVarP(&opts.gitBranch, "git-branch", "", "master",
		"Git branch to be used by Flux")
	cmd.Flags().StringVarP(&opts.gitLabel, "git-label", "", "flux",
		"Git label to keep track of Flux's sync progress; overrides both --git-sync-tag and --git-notes-ref")
	cmd.Flags().StringVarP(&opts.gitUser, "git-user", "", "Flux",
		"Username to use as git committer")
	cmd.Flags().StringVarP(&opts.gitEmail, "git-email", "", "flux-dev@googlegroups.com",
		"Email to use as git committer")
	cmd.Flags().StringVarP(&opts.namespace, "namespace", "", "flux",
		"Cluster namespace where to install flux")
	cmd.Flags().StringSliceVarP(&opts.additionalFluxArgs, "extra-flux-args", "", []string{},
		"Additional arguments for Flux as CSVs, e.g. --extra-flux-arg='--manifest-generation=true,--sync-garbage-collection=true'")
	return cmd
}

func (opts *installOpts) RunE(cmd *cobra.Command, args []string) error {
	//0. Read and patch embedded templates manifests with what was passed to as options

	//1. Print them

	return nil
}
