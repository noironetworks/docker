package service

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/client"
	"github.com/docker/docker/cli"
	"github.com/spf13/cobra"
)

func newScaleCommand(dockerCli *client.DockerCli) *cobra.Command {
	return &cobra.Command{
		Use:   "scale SERVICE=SCALE [SERVICE=SCALE...]",
		Short: "Scale one or multiple services",
		Args:  scaleArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScale(dockerCli, args)
		},
	}
}

func scaleArgs(cmd *cobra.Command, args []string) error {
	if err := cli.RequiresMinArgs(1)(cmd, args); err != nil {
		return err
	}
	for _, arg := range args {
		if parts := strings.SplitN(arg, "=", 2); len(parts) != 2 {
			return fmt.Errorf(
				"Invalid scale specifier '%s'.\nSee '%s --help'.\n\nUsage:  %s\n\n%s",
				arg,
				cmd.CommandPath(),
				cmd.UseLine(),
				cmd.Short,
			)
		}
	}
	return nil
}

func runScale(dockerCli *client.DockerCli, args []string) error {
	var errors []string
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		serviceID, scale := parts[0], parts[1]
		if err := runServiceScale(dockerCli, serviceID, scale); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", serviceID, err.Error()))
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(errors, "\n"))
}

func runServiceScale(dockerCli *client.DockerCli, serviceID string, scale string) error {
	client := dockerCli.Client()
	ctx := context.Background()

	service, err := client.ServiceInspect(ctx, serviceID)
	if err != nil {
		return err
	}

	serviceMode := &service.Spec.Mode
	if serviceMode.Replicated == nil {
		return fmt.Errorf("scale can only be used with replicated mode")
	}
	uintScale, err := strconv.ParseUint(scale, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid scale value %s: %s", scale, err.Error())
	}
	serviceMode.Replicated.Instances = &uintScale

	err = client.ServiceUpdate(ctx, service.ID, service.Version, service.Spec)
	if err != nil {
		return err
	}

	fmt.Fprintf(dockerCli.Out(), "%s scaled to %s", serviceID, scale)
	return nil
}
