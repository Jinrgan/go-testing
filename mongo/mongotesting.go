package mongotesting

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

const (
	image         = "mongo:4.4"
	containerPort = "27017/tcp"
)

var mongoURI string

const defaultMongoURI = "mongodb://localhost:27017"

//RunWithMongoInDocker runs the tests with
// a mongo instance in a docker container.
func RunWithMongoInDocker(m *testing.M) int {
	c, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	resp, err := c.ContainerCreate(ctx, &container.Config{
		Image: image,
		ExposedPorts: nat.PortSet{
			containerPort: {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{
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

	hostPort := inspRes.NetworkSettings.Ports[containerPort][0]
	mongoURI = fmt.Sprintf("mongodb://%s:%s", hostPort.HostIP, hostPort.HostPort)

	return m.Run()
}

//NewClient creates a client connected to the mongo instance in docker.
func NewClient(c context.Context) (*mongo.Client, error) {
	if mongoURI == "" {
		return nil, fmt.Errorf("mongo uri not set, Please run RunWithMongoInDocker in TestMain")
	}

	return mongo.Connect(c, options.Client().ApplyURI(mongoURI))
}

//NewDefaultClient creates a client connected to localhost:27017
func NewDefaultClient(c context.Context) (*mongo.Client, error) {
	return mongo.Connect(c, options.Client().ApplyURI(defaultMongoURI))
}
