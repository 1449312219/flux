package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

type serviceShowOpts struct {
	*serviceOpts
	service string
	limit   int
}

func newServiceShow(parent *serviceOpts) *serviceShowOpts {
	return &serviceShowOpts{serviceOpts: parent}
}

func (opts *serviceShowOpts) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-images",
		Short:   "Show the deployed and available images for a service.",
		Example: makeExample("fluxctl list-images --service=default/foo"),
		RunE:    opts.RunE,
	}
	cmd.Flags().StringVarP(&opts.service, "service", "s", "", "Show images for this service")
	cmd.Flags().IntVarP(&opts.limit, "limit", "n", 10, "Number of images to show (0 for all)")
	return cmd
}

func (opts *serviceShowOpts) RunE(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errorWantedNoArgs
	}

	service, err := parseServiceOption(opts.service)
	if err != nil {
		return err
	}

	services, err := opts.Fluxd.ListImages(service)
	if err != nil {
		return err
	}

	out := newTabwriter()

	fmt.Fprintln(out, "SERVICE\tCONTAINER\tIMAGE\tCREATED")
	for _, service := range services {
		if len(service.Containers) == 0 {
			fmt.Fprintf(out, "%s\t\t\t\n", service.ID)
			continue
		}

		serviceName := service.ID
		var lineCount int
		for _, container := range service.Containers {
			containerName := container.Name
			reg, repo, _ := container.Current.ID.Components()
			if reg != "" {
				reg += "/"
			}
			fmt.Fprintf(out, "%s\t%s\t%s%s\t\n", serviceName, containerName, reg, repo)
			foundRunning := false
			for _, available := range container.Available {
				running := "|  "
				if container.Current.ID == available.ID {
					running = "'->"
					foundRunning = true
				} else if foundRunning {
					running = "   "
				}

				_, _, tag := available.ID.Components()

				lineCount++
				var printElipsis, printLine bool
				if opts.limit <= 0 || lineCount <= opts.limit {
					printElipsis, printLine = false, true
				} else if container.Current.ID == available.ID {
					printElipsis, printLine = lineCount > (opts.limit+1), true
				}
				if printElipsis {
					fmt.Fprintf(out, "\t\t%s\t\n", ":")
				}
				if printLine {
					fmt.Fprintf(out, "\t\t%s %s\t%s\n", running, tag, available.CreatedAt.Format(time.RFC822))
				}
			}
			serviceName = ""
		}
	}
	out.Flush()
	return nil
}
