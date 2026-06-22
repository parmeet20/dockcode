package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Client wraps the official Docker SDK Client to provide a simplified API.
type Client struct {
	cli *client.Client
}

// Container represents container info tailored for DockerCode TUI and tools.
type Container struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	Status  string   `json:"status"`
	Ports   []string `json:"ports"`
	Created int64    `json:"created"`
}

// Image represents image info tailored for DockerCode TUI and tools.
type Image struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Size       int64  `json:"size"`
	Created    int64  `json:"created"`
}

// Volume represents volume info tailored for DockerCode TUI and tools.
type Volume struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Scope      string `json:"scope"`
	Mountpoint string `json:"mountpoint"`
}

// Network represents network info tailored for DockerCode TUI and tools.
type Network struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Driver string `json:"driver"`
	Scope  string `json:"scope"`
}

// NewClient initializes and returns a Client wrapper.
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

// Close closes the underlying Docker client connection.
func (c *Client) Close() error {
	return c.cli.Close()
}

// Ping checks if the Docker daemon is reachable.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

// ListContainers returns list of containers.
func (c *Client) ListContainers(ctx context.Context, all bool) ([]Container, error) {
	rawContainers, err := c.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, err
	}

	containers := make([]Container, 0, len(rawContainers))
	for _, rc := range rawContainers {
		name := ""
		if len(rc.Names) > 0 {
			name = strings.TrimPrefix(rc.Names[0], "/")
		}

		ports := make([]string, 0)
		for _, p := range rc.Ports {
			if p.PublicPort > 0 {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}

		containers = append(containers, Container{
			ID:      rc.ID[:12],
			Name:    name,
			Image:   rc.Image,
			Status:  rc.Status,
			Ports:   ports,
			Created: rc.Created,
		})
	}
	return containers, nil
}

// ListImages returns list of local images.
func (c *Client) ListImages(ctx context.Context) ([]Image, error) {
	rawImages, err := c.cli.ImageList(ctx, image.ListOptions{All: false})
	if err != nil {
		return nil, err
	}

	images := make([]Image, 0, len(rawImages))
	for _, ri := range rawImages {
		repo := "<none>"
		tag := "<none>"
		if len(ri.RepoTags) > 0 {
			parts := strings.SplitN(ri.RepoTags[0], ":", 2)
			repo = parts[0]
			if len(parts) > 1 {
				tag = parts[1]
			}
		}

		id := ri.ID
		if strings.HasPrefix(id, "sha256:") && len(id) > 19 {
			id = id[7:19]
		}

		images = append(images, Image{
			ID:         id,
			Repository: repo,
			Tag:        tag,
			Size:       ri.Size,
			Created:    ri.Created,
		})
	}
	return images, nil
}

// ListVolumes returns list of local volumes.
func (c *Client) ListVolumes(ctx context.Context) ([]Volume, error) {
	volList, err := c.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	volumes := make([]Volume, 0, len(volList.Volumes))
	for _, rv := range volList.Volumes {
		volumes = append(volumes, Volume{
			Name:       rv.Name,
			Driver:     rv.Driver,
			Scope:      rv.Scope,
			Mountpoint: rv.Mountpoint,
		})
	}
	return volumes, nil
}

// ListNetworks returns list of local networks.
func (c *Client) ListNetworks(ctx context.Context) ([]Network, error) {
	rawNetworks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	networks := make([]Network, 0, len(rawNetworks))
	for _, rn := range rawNetworks {
		id := rn.ID
		if len(id) > 12 {
			id = id[:12]
		}
		networks = append(networks, Network{
			ID:     id,
			Name:   rn.Name,
			Driver: rn.Driver,
			Scope:  rn.Scope,
		})
	}
	return networks, nil
}

// PullImage pulls a Docker image and blocks until it is complete.
func (c *Client) PullImage(ctx context.Context, imageName string) (int64, error) {
	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	// Consume entire pull stream to complete operation
	_, _ = io.Copy(io.Discard, reader)

	// Fetch the newly pulled image to find its size
	inspect, _, err := c.cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return 0, nil // return success since pull finished, but size is unknown
	}
	return inspect.Size, nil
}

// RemoveImage removes a local image.
func (c *Client) RemoveImage(ctx context.Context, imageID string, force bool) error {
	_, err := c.cli.ImageRemove(ctx, imageID, image.RemoveOptions{
		Force:         force,
		PruneChildren: true,
	})
	return err
}

// BuildImage builds an image from a context reader (e.g. tarball) and writes output to the writer.
func (c *Client) BuildImage(ctx context.Context, buildContext io.Reader, tag string, output io.Writer) error {
	resp, err := c.cli.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: "Dockerfile",
		Remove:     true,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	return err
}

// StartContainer starts a created container.
func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

// StopContainer stops a running container.
func (c *Client) StopContainer(ctx context.Context, name string) error {
	return c.cli.ContainerStop(ctx, name, container.StopOptions{})
}

// RemoveContainer removes a container.
func (c *Client) RemoveContainer(ctx context.Context, name string, force bool) error {
	return c.cli.ContainerRemove(ctx, name, container.RemoveOptions{
		Force:         force,
		RemoveVolumes: true,
	})
}

// GetContainerLogs returns logs of a container as a string.
func (c *Client) GetContainerLogs(ctx context.Context, name string, tail int) (string, error) {
	reader, err := c.cli.ContainerLogs(ctx, name, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
	})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	return buf.String(), nil
}

// RunContainer runs a container by pulling the image if needed, creating, and starting it.
func (c *Client) RunContainer(ctx context.Context, image string, name string, ports []string, env map[string]string, volumes []string, restart string, detach bool) (string, error) {
	// 1. Pull the image if it does not exist locally
	_, _ = c.PullImage(ctx, image)

	// Convert ports (e.g. ["80:80"])
	portMap := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for _, p := range ports {
		parts := strings.SplitN(p, ":", 2)
		var hostPort, containerPort string
		if len(parts) == 2 {
			hostPort = parts[0]
			containerPort = parts[1]
		} else {
			containerPort = parts[0]
		}

		cPort, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return "", fmt.Errorf("invalid port format: %w", err)
		}
		exposedPorts[cPort] = struct{}{}
		if hostPort != "" {
			portMap[cPort] = []nat.PortBinding{{HostPort: hostPort}}
		}
	}

	// Env list
	envList := make([]string, 0, len(env))
	for k, v := range env {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}

	// Restart policy
	restartPolicy := container.RestartPolicyMode(restart)
	if restartPolicy == "" {
		restartPolicy = container.RestartPolicyDisabled
	}

	config := &container.Config{
		Image:        image,
		ExposedPorts: exposedPorts,
		Env:          envList,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portMap,
		Binds:        volumes,
		RestartPolicy: container.RestartPolicy{
			Name: restartPolicy,
		},
	}

	created, err := c.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", err
	}

	err = c.cli.ContainerStart(ctx, created.ID, container.StartOptions{})
	if err != nil {
		return "", err
	}

	return created.ID, nil
}

// ExecCommand runs a command inside a running container and returns stdout+stderr combined.
func (c *Client) ExecCommand(ctx context.Context, name string, cmd string) (string, error) {
	args := strings.Fields(cmd)
	if len(args) == 0 {
		return "", fmt.Errorf("empty command")
	}

	execCreate, err := c.cli.ContainerExecCreate(ctx, name, container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          args,
	})
	if err != nil {
		return "", err
	}

	resp, err := c.cli.ContainerExecAttach(ctx, execCreate.ID, container.ExecStartOptions{})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Reader)
	return buf.String(), nil
}
