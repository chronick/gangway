package translate

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AppleContainer is the JSON structure output by `container list --format json`.
type AppleContainer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	Command   string `json:"command"`
}

// AppleContainerInspect is the JSON structure from `container inspect --format json`.
type AppleContainerInspect struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Status    string `json:"status"`
	Pid       int    `json:"pid"`
	ExitCode  int    `json:"exitCode"`
	CreatedAt string `json:"createdAt"`
	StartedAt string `json:"startedAt"`
	StoppedAt string `json:"stoppedAt"`
	Command   string `json:"command"`
	Env       []string `json:"env"`
}

// AppleImage is the JSON structure from `container image list --format json`.
type AppleImage struct {
	ID        string   `json:"id"`
	Tags      []string `json:"tags"`
	Digests   []string `json:"digests"`
	CreatedAt string   `json:"createdAt"`
	Size      int64    `json:"size"`
}

// AppleNetwork is the JSON structure from `container network list --format json`.
type AppleNetwork struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

// ParseContainerList parses `container list --format json` output into Docker API format.
func ParseContainerList(output string) ([]ContainerListEntry, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return []ContainerListEntry{}, nil
	}

	var appleContainers []AppleContainer
	if err := json.Unmarshal([]byte(output), &appleContainers); err != nil {
		// Try parsing as single object.
		var single AppleContainer
		if err2 := json.Unmarshal([]byte(output), &single); err2 != nil {
			return nil, fmt.Errorf("parse container list: %w", err)
		}
		appleContainers = []AppleContainer{single}
	}

	result := make([]ContainerListEntry, 0, len(appleContainers))
	for _, ac := range appleContainers {
		entry := ContainerListEntry{
			ID:      ac.ID,
			Names:   []string{"/" + ac.Name},
			Image:   ac.Image,
			ImageID: "",
			Command: ac.Command,
			Created: parseTimeUnix(ac.CreatedAt),
			State:   normalizeState(ac.Status),
			Status:  ac.Status,
			Ports:   []Port{},
			Labels:  map[string]string{},
		}
		result = append(result, entry)
	}
	return result, nil
}

// ParseContainerInspect parses `container inspect --format json` output into Docker API format.
func ParseContainerInspect(output string) (*ContainerJSON, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("empty inspect output")
	}

	var ai AppleContainerInspect
	if err := json.Unmarshal([]byte(output), &ai); err != nil {
		// Try as array (some versions wrap in array).
		var arr []AppleContainerInspect
		if err2 := json.Unmarshal([]byte(output), &arr); err2 != nil {
			return nil, fmt.Errorf("parse container inspect: %w", err)
		}
		if len(arr) == 0 {
			return nil, fmt.Errorf("empty inspect array")
		}
		ai = arr[0]
	}

	running := normalizeState(ai.Status) == "running"

	result := &ContainerJSON{
		ID:      ai.ID,
		Created: ai.CreatedAt,
		Name:    "/" + ai.Name,
		Image:   ai.Image,
		State: ContainerState{
			Status:     normalizeState(ai.Status),
			Running:    running,
			Paused:     false,
			Pid:        ai.Pid,
			ExitCode:   ai.ExitCode,
			StartedAt:  ai.StartedAt,
			FinishedAt: ai.StoppedAt,
		},
		Config: ContainerConfig{
			Image:  ai.Image,
			Cmd:    parseCommand(ai.Command),
			Env:    ai.Env,
			Labels: map[string]string{},
		},
	}
	return result, nil
}

// ParseImageList parses `container image list --format json` output into Docker API format.
func ParseImageList(output string) ([]ImageListEntry, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return []ImageListEntry{}, nil
	}

	var appleImages []AppleImage
	if err := json.Unmarshal([]byte(output), &appleImages); err != nil {
		var single AppleImage
		if err2 := json.Unmarshal([]byte(output), &single); err2 != nil {
			return nil, fmt.Errorf("parse image list: %w", err)
		}
		appleImages = []AppleImage{single}
	}

	result := make([]ImageListEntry, 0, len(appleImages))
	for _, ai := range appleImages {
		entry := ImageListEntry{
			ID:          ai.ID,
			RepoTags:    ai.Tags,
			RepoDigests: ai.Digests,
			Created:     parseTimeUnix(ai.CreatedAt),
			Size:        ai.Size,
		}
		if entry.RepoTags == nil {
			entry.RepoTags = []string{}
		}
		if entry.RepoDigests == nil {
			entry.RepoDigests = []string{}
		}
		result = append(result, entry)
	}
	return result, nil
}

// ParseNetworkList parses `container network list --format json` output into Docker API format.
func ParseNetworkList(output string) ([]NetworkListEntry, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return []NetworkListEntry{}, nil
	}

	var appleNetworks []AppleNetwork
	if err := json.Unmarshal([]byte(output), &appleNetworks); err != nil {
		var single AppleNetwork
		if err2 := json.Unmarshal([]byte(output), &single); err2 != nil {
			return nil, fmt.Errorf("parse network list: %w", err)
		}
		appleNetworks = []AppleNetwork{single}
	}

	result := make([]NetworkListEntry, 0, len(appleNetworks))
	for _, an := range appleNetworks {
		entry := NetworkListEntry{
			ID:     an.ID,
			Name:   an.Name,
			Driver: an.Driver,
			Scope:  "local",
			Labels: map[string]string{},
		}
		result = append(result, entry)
	}
	return result, nil
}

// normalizeState maps Apple container status strings to Docker state strings.
func normalizeState(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	switch {
	case s == "running":
		return "running"
	case s == "stopped", s == "exited":
		return "exited"
	case s == "created":
		return "created"
	case s == "paused":
		return "paused"
	case strings.Contains(s, "running"):
		return "running"
	case strings.Contains(s, "stop"), strings.Contains(s, "exit"):
		return "exited"
	default:
		return s
	}
}

// parseTimeUnix parses a time string and returns Unix timestamp.
// Returns 0 if parsing fails.
func parseTimeUnix(s string) int64 {
	if s == "" {
		return 0
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t.Unix()
		}
	}
	return 0
}

// parseCommand splits a command string into parts.
func parseCommand(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}
	return strings.Fields(cmd)
}
