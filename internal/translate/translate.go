package translate

import (
	"fmt"
	"strings"
)

// CreateArgs translates a Docker create container request into container CLI arguments.
// Returns the args to pass to `container run` (without the binary name).
func CreateArgs(req CreateContainerRequest, name string) []string {
	args := []string{"run"}

	if name != "" {
		args = append(args, "--name", name)
	}

	// Environment variables
	for _, env := range req.Env {
		args = append(args, "--env", env)
	}

	// Volume binds
	if req.HostConfig != nil {
		for _, bind := range req.HostConfig.Binds {
			args = append(args, "--volume", bind)
		}

		// Port bindings
		for containerPort, bindings := range req.HostConfig.PortBindings {
			for _, binding := range bindings {
				port := strings.TrimSuffix(containerPort, "/tcp")
				port = strings.TrimSuffix(port, "/udp")
				if binding.HostPort != "" {
					args = append(args, "--publish", fmt.Sprintf("%s:%s", binding.HostPort, port))
				} else {
					args = append(args, "--publish", port)
				}
			}
		}

		// Network mode
		if req.HostConfig.NetworkMode != "" {
			args = append(args, "--network", req.HostConfig.NetworkMode)
		}
	}

	// Labels
	for k, v := range req.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
	}

	// Image (required)
	args = append(args, req.Image)

	// Command
	args = append(args, req.Cmd...)

	return args
}

// StopArgs returns args for stopping a container.
func StopArgs(id string, timeout int) []string {
	if timeout > 0 {
		return []string{"stop", "--time", fmt.Sprintf("%d", timeout), id}
	}
	return []string{"stop", id}
}

// KillArgs returns args for killing a container.
func KillArgs(id string, signal string) []string {
	if signal != "" {
		return []string{"kill", "--signal", signal, id}
	}
	return []string{"kill", id}
}

// RemoveArgs returns args for removing a container.
func RemoveArgs(id string, force bool) []string {
	if force {
		return []string{"delete", "--force", id}
	}
	return []string{"delete", id}
}

// StartArgs returns args for starting a container.
func StartArgs(id string) []string {
	return []string{"start", id}
}

// InspectArgs returns args for inspecting a container.
func InspectArgs(id string) []string {
	return []string{"inspect", "--format", "json", id}
}

// ListArgs returns args for listing containers.
func ListArgs(all bool) []string {
	args := []string{"list", "--format", "json"}
	if all {
		args = append(args, "--all")
	}
	return args
}

// LogsArgs returns args for getting container logs.
func LogsArgs(id string, follow bool, tail string, timestamps bool) []string {
	args := []string{"logs"}
	if follow {
		args = append(args, "--follow")
	}
	if tail != "" && tail != "all" {
		args = append(args, "--tail", tail)
	}
	if timestamps {
		args = append(args, "--timestamps")
	}
	args = append(args, id)
	return args
}

// ExecArgs returns args for executing a command in a container.
func ExecArgs(id string, cmd []string) []string {
	args := []string{"exec", id}
	args = append(args, cmd...)
	return args
}

// ImageListArgs returns args for listing images.
func ImageListArgs() []string {
	return []string{"image", "list", "--format", "json"}
}

// ImagePullArgs returns args for pulling an image.
func ImagePullArgs(name string) []string {
	return []string{"image", "pull", name}
}

// NetworkCreateArgs returns args for creating a network.
func NetworkCreateArgs(req NetworkCreateRequest) []string {
	args := []string{"network", "create"}
	if req.Driver != "" {
		args = append(args, "--driver", req.Driver)
	}
	args = append(args, req.Name)
	return args
}

// NetworkDeleteArgs returns args for deleting a network.
func NetworkDeleteArgs(id string) []string {
	return []string{"network", "delete", id}
}

// NetworkListArgs returns args for listing networks.
func NetworkListArgs() []string {
	return []string{"network", "list", "--format", "json"}
}
