package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type repoImagesOpts struct {
	*repoOpts
}

func newRepoImages(parent *repoOpts) *repoImagesOpts {
	return &repoImagesOpts{repoOpts: parent}
}

func (opts *repoImagesOpts) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "List images available in an image repository.",
		RunE:  opts.RunE,
	}
	return cmd
}

func (opts *repoImagesOpts) RunE(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errorWantedNoArgs
	}
	if opts.repository == "" {
		return fmt.Errorf("expected flag --repository, giving the repository for which to list images")
	}

	images, err := opts.Fluxd.Images(opts.repository)
	if err != nil {
		return err
	}

	out := newTabWriter()
	fmt.Fprintln(out, "IMAGE\tCREATED")
	for _, image := range images {
		fmt.Fprintf(out, "%s:%s\t%s\n", image.Name, image.Tag, image.CreatedAt)
	}
	out.Flush()
	return nil
}
