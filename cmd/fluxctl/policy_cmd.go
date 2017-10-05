package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/policy"
	"github.com/weaveworks/flux/update"
)

type servicePolicyOpts struct {
	*rootOpts
	outputOpts

	service string
	tagAll  string
	tags    []string

	automate, deautomate bool
	lock, unlock         bool

	cause update.Cause
}

func newServicePolicy(parent *rootOpts) *servicePolicyOpts {
	return &servicePolicyOpts{rootOpts: parent}
}

func (opts *servicePolicyOpts) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage policies for a service.",
		Long: `
Manage policies for a service.

Tag filter patterns must be specified as 'container=pattern', such as 'foo=1.*'
where an asterisk means 'match anything'.
Surrounding these with single-quotes are recommended to avoid shell expansion.

If both --tag-all and --tag are specified, --tag-all will apply to all
containers which aren't explicitly named.
        `,
		Example: makeExample(
			"fluxctl policy --service=foo --automate",
			"fluxctl policy --service=foo --lock",
			"fluxctl policy --service=foo --tag='bar=1.*' --tag='baz=2.*'",
			"fluxctl policy --service=foo --tag-all='master-*' --tag='bar=1.*'",
		),
		RunE: opts.RunE,
	}

	AddOutputFlags(cmd, &opts.outputOpts)
	AddCauseFlags(cmd, &opts.cause)
	flags := cmd.Flags()
	flags.StringVarP(&opts.service, "service", "s", "", "Service to modify")
	flags.StringVar(&opts.tagAll, "tag-all", "", "Tag filter pattern to apply to all containers")
	flags.StringSliceVar(&opts.tags, "tag", nil, "Tag filter container/pattern pairs")
	flags.BoolVar(&opts.automate, "automate", false, "Automate service")
	flags.BoolVar(&opts.deautomate, "deautomate", false, "Deautomate for service")
	flags.BoolVar(&opts.lock, "lock", false, "Lock service")
	flags.BoolVar(&opts.unlock, "unlock", false, "Unlock service")

	return cmd
}

func (opts *servicePolicyOpts) RunE(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errorWantedNoArgs
	}
	if opts.service == "" {
		return newUsageError("-s, --service is required")
	}
	if opts.automate && opts.deautomate {
		return newUsageError("automate and deautomate both specified")
	}
	if opts.lock && opts.unlock {
		return newUsageError("lock and unlock both specified")
	}

	serviceID, err := flux.ParseResourceID(opts.service)
	if err != nil {
		return err
	}

	update, err := calculatePolicyChanges(opts)
	if err != nil {
		return err
	}

	ctx := context.Background()

	jobID, err := opts.API.UpdatePolicies(ctx, policy.Updates{
		serviceID: update,
	}, opts.cause)
	if err != nil {
		return err
	}
	return await(ctx, cmd.OutOrStdout(), cmd.OutOrStderr(), opts.API, jobID, false, opts.verbose)
}

func calculatePolicyChanges(opts *servicePolicyOpts) (policy.Update, error) {
	add := policy.Set{}
	if opts.automate {
		add = add.Add(policy.Automated)
	}
	if opts.lock {
		add = add.Add(policy.Locked)
		if opts.cause.User != "" {
			add = add.
				Set(policy.LockedUser, opts.cause.User).
				Set(policy.LockedMsg, opts.cause.Message)
		}
	}

	remove := policy.Set{}
	if opts.deautomate {
		remove = remove.Add(policy.Automated)
	}
	if opts.unlock {
		remove = remove.
			Add(policy.Locked).
			Add(policy.LockedMsg).
			Add(policy.LockedUser)
	}
	if opts.tagAll != "" {
		add = add.Set(policy.TagAll, "glob:"+opts.tagAll)
	}

	for _, tagPair := range opts.tags {
		parts := strings.Split(tagPair, "=")
		if len(parts) != 2 {
			return policy.Update{}, fmt.Errorf("invalid container/tag pair: %q. Expected format is 'container=filter'", tagPair)
		}

		container, tag := parts[0], parts[1]
		if tag != "*" {
			add = add.Set(policy.TagPrefix(container), "glob:"+tag)
		} else {
			remove = remove.Add(policy.TagPrefix(container))
		}
	}

	return policy.Update{
		Add:    add,
		Remove: remove,
	}, nil
}
