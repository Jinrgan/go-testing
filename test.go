package unit_test

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DBConfig struct {
	Image         string
	ContainerPort string
	Env           []string
	DefaultURI    string                // 默认数据库连接 URI
	ConnFormatter func(ip, port string) // 连接字符串格式模板
}

func RunInDocker(cfg DBConfig, m *testing.M) int {
	c, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	port := nat.Port(cfg.ContainerPort)
	resp, err := c.ContainerCreate(ctx, &container.Config{
		Image: cfg.Image,
		ExposedPorts: nat.PortSet{
			port: {},
		},
		Env: cfg.Env,
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   "127.0.0.1",
					HostPort: "0",
				},
			},
		},
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}
	containerID := resp.ID
	defer func() {
		err = c.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			panic(err)
		}
	}()

	err = c.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	inspRes, err := c.ContainerInspect(ctx, containerID)
	if err != nil {
		panic(err)
	}

	hostPort := inspRes.NetworkSettings.Ports[port][0]
	cfg.ConnFormatter(hostPort.HostIP, hostPort.HostPort)

	return m.Run()
}
