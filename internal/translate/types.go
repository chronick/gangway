package translate

// Docker API response types. These are the subset of Docker Engine API types
// needed by OpenClaw and dockerode.

// ContainerListEntry is a single entry from GET /containers/json.
type ContainerListEntry struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	ImageID string            `json:"ImageID"`
	Command string            `json:"Command"`
	Created int64             `json:"Created"`
	State   string            `json:"State"`
	Status  string            `json:"Status"`
	Ports   []Port            `json:"Ports"`
	Labels  map[string]string `json:"Labels"`
}

// Port represents a container port mapping.
type Port struct {
	IP          string `json:"IP,omitempty"`
	PrivatePort uint16 `json:"PrivatePort"`
	PublicPort  uint16 `json:"PublicPort,omitempty"`
	Type        string `json:"Type"`
}

// ContainerJSON is the full container inspect response.
type ContainerJSON struct {
	ID      string          `json:"Id"`
	Created string          `json:"Created"`
	Name    string          `json:"Name"`
	State   ContainerState  `json:"State"`
	Config  ContainerConfig `json:"Config"`
	Image   string          `json:"Image"`
}

// ContainerState holds runtime state.
type ContainerState struct {
	Status     string `json:"Status"`
	Running    bool   `json:"Running"`
	Paused     bool   `json:"Paused"`
	Pid        int    `json:"Pid"`
	ExitCode   int    `json:"ExitCode"`
	StartedAt  string `json:"StartedAt"`
	FinishedAt string `json:"FinishedAt"`
}

// ContainerConfig holds container configuration.
type ContainerConfig struct {
	Image  string            `json:"Image"`
	Cmd    []string          `json:"Cmd"`
	Env    []string          `json:"Env"`
	Labels map[string]string `json:"Labels"`
}

// CreateContainerRequest is the POST body for /containers/create.
type CreateContainerRequest struct {
	Image      string            `json:"Image"`
	Cmd        []string          `json:"Cmd"`
	Env        []string          `json:"Env"`
	Labels     map[string]string `json:"Labels"`
	HostConfig *HostConfig       `json:"HostConfig,omitempty"`
}

// HostConfig holds host-specific container configuration.
type HostConfig struct {
	Binds        []string      `json:"Binds,omitempty"`
	PortBindings map[string][]PortBinding `json:"PortBindings,omitempty"`
	NetworkMode  string        `json:"NetworkMode,omitempty"`
}

// PortBinding represents a port binding.
type PortBinding struct {
	HostIP   string `json:"HostIp,omitempty"`
	HostPort string `json:"HostPort,omitempty"`
}

// CreateContainerResponse is the response body for POST /containers/create.
type CreateContainerResponse struct {
	ID       string   `json:"Id"`
	Warnings []string `json:"Warnings"`
}

// VersionResponse is the response for GET /version.
type VersionResponse struct {
	Version       string `json:"Version"`
	APIVersion    string `json:"ApiVersion"`
	MinAPIVersion string `json:"MinAPIVersion"`
	GitCommit     string `json:"GitCommit"`
	GoVersion     string `json:"GoVersion"`
	Os            string `json:"Os"`
	Arch          string `json:"Arch"`
	KernelVersion string `json:"KernelVersion"`
}

// InfoResponse is the response for GET /info.
type InfoResponse struct {
	ID                string `json:"ID"`
	Containers        int    `json:"Containers"`
	ContainersRunning int    `json:"ContainersRunning"`
	ContainersPaused  int    `json:"ContainersPaused"`
	ContainersStopped int    `json:"ContainersStopped"`
	Images            int    `json:"Images"`
	Driver            string `json:"Driver"`
	OperatingSystem   string `json:"OperatingSystem"`
	OSType            string `json:"OSType"`
	Architecture      string `json:"Architecture"`
	Name              string `json:"Name"`
	ServerVersion     string `json:"ServerVersion"`
}

// ImageListEntry is a single entry from GET /images/json.
type ImageListEntry struct {
	ID          string   `json:"Id"`
	RepoTags    []string `json:"RepoTags"`
	RepoDigests []string `json:"RepoDigests"`
	Created     int64    `json:"Created"`
	Size        int64    `json:"Size"`
}

// NetworkListEntry is a single entry from GET /networks.
type NetworkListEntry struct {
	ID     string            `json:"Id"`
	Name   string            `json:"Name"`
	Driver string            `json:"Driver"`
	Scope  string            `json:"Scope"`
	Labels map[string]string `json:"Labels"`
}

// NetworkCreateRequest is the POST body for /networks/create.
type NetworkCreateRequest struct {
	Name   string            `json:"Name"`
	Driver string            `json:"Driver,omitempty"`
	Labels map[string]string `json:"Labels,omitempty"`
}

// NetworkCreateResponse is the response for POST /networks/create.
type NetworkCreateResponse struct {
	ID      string `json:"Id"`
	Warning string `json:"Warning"`
}

// ExecCreateRequest is the POST body for /containers/{id}/exec.
type ExecCreateRequest struct {
	AttachStdin  bool     `json:"AttachStdin"`
	AttachStdout bool     `json:"AttachStdout"`
	AttachStderr bool     `json:"AttachStderr"`
	Tty          bool     `json:"Tty"`
	Cmd          []string `json:"Cmd"`
}

// ExecCreateResponse is the response for POST /containers/{id}/exec.
type ExecCreateResponse struct {
	ID string `json:"Id"`
}

// ExecStartRequest is the POST body for /exec/{id}/start.
type ExecStartRequest struct {
	Detach bool `json:"Detach"`
	Tty    bool `json:"Tty"`
}
