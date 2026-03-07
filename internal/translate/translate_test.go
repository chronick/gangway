package translate

import (
	"reflect"
	"testing"
)

func TestCreateArgs_Basic(t *testing.T) {
	req := CreateContainerRequest{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "hello"},
	}
	args := CreateArgs(req, "mycontainer")
	expected := []string{"run", "--name", "mycontainer", "alpine:latest", "echo", "hello"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestCreateArgs_WithEnv(t *testing.T) {
	req := CreateContainerRequest{
		Image: "nginx",
		Env:   []string{"FOO=bar", "BAZ=qux"},
	}
	args := CreateArgs(req, "")

	// Should contain --env flags
	found := 0
	for i, a := range args {
		if a == "--env" {
			found++
			if i+1 >= len(args) {
				t.Fatal("--env without value")
			}
		}
	}
	if found != 2 {
		t.Errorf("expected 2 --env flags, got %d", found)
	}
}

func TestCreateArgs_WithVolumes(t *testing.T) {
	req := CreateContainerRequest{
		Image: "nginx",
		HostConfig: &HostConfig{
			Binds: []string{"/host/path:/container/path", "/data:/data:ro"},
		},
	}
	args := CreateArgs(req, "vol-test")

	found := 0
	for _, a := range args {
		if a == "--volume" {
			found++
		}
	}
	if found != 2 {
		t.Errorf("expected 2 --volume flags, got %d", found)
	}
}

func TestCreateArgs_WithPorts(t *testing.T) {
	req := CreateContainerRequest{
		Image: "nginx",
		HostConfig: &HostConfig{
			PortBindings: map[string][]PortBinding{
				"80/tcp": {{HostPort: "8080"}},
			},
		},
	}
	args := CreateArgs(req, "port-test")

	foundPublish := false
	for i, a := range args {
		if a == "--publish" && i+1 < len(args) && args[i+1] == "8080:80" {
			foundPublish = true
		}
	}
	if !foundPublish {
		t.Errorf("expected --publish 8080:80 in args %v", args)
	}
}

func TestCreateArgs_WithNetwork(t *testing.T) {
	req := CreateContainerRequest{
		Image: "nginx",
		HostConfig: &HostConfig{
			NetworkMode: "mynet",
		},
	}
	args := CreateArgs(req, "")

	foundNetwork := false
	for i, a := range args {
		if a == "--network" && i+1 < len(args) && args[i+1] == "mynet" {
			foundNetwork = true
		}
	}
	if !foundNetwork {
		t.Errorf("expected --network mynet in args %v", args)
	}
}

func TestCreateArgs_NoName(t *testing.T) {
	req := CreateContainerRequest{
		Image: "alpine",
	}
	args := CreateArgs(req, "")

	for _, a := range args {
		if a == "--name" {
			t.Error("should not include --name when name is empty")
		}
	}
}

func TestStopArgs(t *testing.T) {
	args := StopArgs("abc123", 0)
	if !reflect.DeepEqual(args, []string{"stop", "abc123"}) {
		t.Errorf("unexpected args: %v", args)
	}

	args = StopArgs("abc123", 30)
	if !reflect.DeepEqual(args, []string{"stop", "--time", "30", "abc123"}) {
		t.Errorf("unexpected args with timeout: %v", args)
	}
}

func TestKillArgs(t *testing.T) {
	args := KillArgs("abc123", "")
	if !reflect.DeepEqual(args, []string{"kill", "abc123"}) {
		t.Errorf("unexpected args: %v", args)
	}

	args = KillArgs("abc123", "SIGKILL")
	if !reflect.DeepEqual(args, []string{"kill", "--signal", "SIGKILL", "abc123"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestRemoveArgs(t *testing.T) {
	args := RemoveArgs("abc123", false)
	if !reflect.DeepEqual(args, []string{"delete", "abc123"}) {
		t.Errorf("unexpected args: %v", args)
	}

	args = RemoveArgs("abc123", true)
	if !reflect.DeepEqual(args, []string{"delete", "--force", "abc123"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestStartArgs(t *testing.T) {
	args := StartArgs("abc123")
	if !reflect.DeepEqual(args, []string{"start", "abc123"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestInspectArgs(t *testing.T) {
	args := InspectArgs("abc123")
	if !reflect.DeepEqual(args, []string{"inspect", "--format", "json", "abc123"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestListArgs(t *testing.T) {
	args := ListArgs(false)
	if !reflect.DeepEqual(args, []string{"list", "--format", "json"}) {
		t.Errorf("unexpected args: %v", args)
	}

	args = ListArgs(true)
	if !reflect.DeepEqual(args, []string{"list", "--format", "json", "--all"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestLogsArgs(t *testing.T) {
	args := LogsArgs("abc123", false, "", false)
	if !reflect.DeepEqual(args, []string{"logs", "abc123"}) {
		t.Errorf("unexpected args: %v", args)
	}

	args = LogsArgs("abc123", true, "100", true)
	expected := []string{"logs", "--follow", "--tail", "100", "--timestamps", "abc123"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestExecArgs(t *testing.T) {
	args := ExecArgs("abc123", []string{"sh", "-c", "echo hello"})
	expected := []string{"exec", "abc123", "sh", "-c", "echo hello"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestImageListArgs(t *testing.T) {
	args := ImageListArgs()
	if !reflect.DeepEqual(args, []string{"image", "list", "--format", "json"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestImagePullArgs(t *testing.T) {
	args := ImagePullArgs("alpine:latest")
	if !reflect.DeepEqual(args, []string{"image", "pull", "alpine:latest"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestNetworkCreateArgs(t *testing.T) {
	req := NetworkCreateRequest{Name: "mynet", Driver: "bridge"}
	args := NetworkCreateArgs(req)
	if !reflect.DeepEqual(args, []string{"network", "create", "--driver", "bridge", "mynet"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestNetworkDeleteArgs(t *testing.T) {
	args := NetworkDeleteArgs("mynet")
	if !reflect.DeepEqual(args, []string{"network", "delete", "mynet"}) {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestNetworkListArgs(t *testing.T) {
	args := NetworkListArgs()
	if !reflect.DeepEqual(args, []string{"network", "list", "--format", "json"}) {
		t.Errorf("unexpected args: %v", args)
	}
}
