# Flux

We believe in GitOps:

- **You declaratively describe the entire desired state of your
  system in git.** This includes the apps, config, dashboards,
  monitoring and everything else.
- **What can be described can be automated.** Use YAMLs to enforce
  conformance of the system. You don't need to run `kubectl`, all changes go
  through git. This allows you to diff against the observed
  state.
- **You push code not containers.** Everything is controlled through
  pull requests. Instant deployments. No learning curve for new devs, just
  use git. This allows you to recover from any snapshot as you have
  an atomic sequence of transactions.

Flux is a tool that automatically ensures that the state of a cluster
matches the config in git. It uses an operator in the cluster to trigger
deployments inside Kubernetes, which means you don't need a separate CD tool,
because it's cloud-native. It monitors all relevant image repositories, detects
new images, triggers deployments and updates the desired running configuration
based on that (and a configurable policy).

The benefits are: you don't need to grant your CI access to the cluster, every
change is atomic and transactional, git has your audit log. Each transaction
either fails or succeeds cleanly. You're entirely code centric and don't new
infrastructure.

![Deployment Pipeline](site/images/deployment-pipeline.png)

[![CircleCI](https://circleci.com/gh/weaveworks/flux.svg?style=svg)](https://circleci.com/gh/weaveworks/flux)
[![GoDoc](https://godoc.org/github.com/weaveworks/flux?status.svg)](https://godoc.org/github.com/weaveworks/flux)

## GitOps

Git has moved the state of the art forward in development. A decade
of best practices says that config is code, and code should be stored
in version control. Now it is paying that benefit forward to Ops. It
is much more transparent to fix a production issue via a pull request,
than to make changes to the running system.

At its core the GitOps pattern encourages you to

- Make all provisioning and deployment configuration declarative
- Keep the entire system state under version control and described in
  a single Git repository
- Make operational changes by pull request (plus build & release pipelines)
- Let diff tools detect any divergence and notify you; and use
  sync tools enable convergence
- Get audit logs via Git

## What Flux does

Flux is most useful when used as a deployment tool at the end of a
Continuous Delivery pipeline. Flux will make sure that your new
container images and config changes are propagated to the cluster.

Among its features are:

- [Automated git → cluster synchronisation](/site/introduction.md#automated-git-cluster-synchronisation)
- [Automated deployment of new containers](/site/introduction.md#automated-deployment-of-new-containers)
- [Integrations with other devops tools](/site/introduction.md#integrations-with-other-devops-tools) ([Helm](site/helm/helm-integration.md) and more)
- No additional service or infrastructure needed - Flux lives inside your
  cluster

## Get started with Flux

Get started by browsing through the documentation below:

- [Introduction to Flux](/site/introduction.md)
- [FAQ](/site/faq.md)
- [How it works](/site/how-it-works.md)
- [Installing Flux](/site/installing.md)
- [Using Flux](/site/using.md)
- [Upgrading to Flux v1](/site/upgrading-to-1.0.md)
- [Troubleshooting](/site/troubleshooting.md)

## Developer information

[Build documentation](/site/building.md)

[Release documentation](/internal_docs/releasing.md)

### Contribution

Flux follows a typical PR workflow.
All contributions should be made as PRs that satisfy the guidelines below.

### Guidelines

- All code must abide [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Names should abide [What's in a name](https://talks.golang.org/2014/names.slide#1)
- Code must build on both Linux and Darwin, via plain `go build`
- Code should have appropriate test coverage, invoked via plain `go test`

In addition, several mechanical checks are enforced.
See [the lint script](/lint) for details.

## <a name="help"></a>Getting Help

If you have any questions about Flux and continuous delivery:

- Read [the Weave Flux docs](https://github.com/weaveworks/flux/tree/master/site).
- Invite yourself to the <a href="https://weaveworks.github.io/community-slack/" target="_blank">Weave community</a> slack.
- Ask a question on the [#flux](https://weave-community.slack.com/messages/flux/) slack channel.
- Join the <a href="https://www.meetup.com/pro/Weave/"> Weave User Group </a> and get invited to online talks, hands-on training and meetups in your area.
- Send an email to <a href="mailto:weave-users@weave.works">weave-users@weave.works</a>
- <a href="https://github.com/weaveworks/flux/issues/new">File an issue.</a>

Your feedback is always welcome!
