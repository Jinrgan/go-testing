package mysqltesting

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	image         = "mysql:5.6"
	containerPort = "3306/tcp"
	pwdEnv        = "MYSQL_ROOT_PASSWORD=root"
	dbEnv         = "MYSQL_DATABASE=test"
)

var mysqlDSN string

//RunWithMysqlInDocker runs the tests with
// a mysql instance in a docker container.
func RunWithMysqlInDocker(m *testing.M) int {
	clt, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	resp, err := clt.ContainerCreate(
		ctx,
		&container.Config{
			ExposedPorts: nat.PortSet{
				containerPort: {},
			},
			Env:   []string{pwdEnv, dbEnv},
			Image: image,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				containerPort: []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "0",
					},
				},
			},
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		panic(err)
	}
	containerID := resp.ID
	defer func() {
		err := clt.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			panic(err)
		}
	}()

	err = clt.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	inspRes, err := clt.ContainerInspect(ctx, containerID)
	if err != nil {
		panic(err)
	}

	hostPort := inspRes.NetworkSettings.Ports[containerPort][0]

	mysqlDSN = fmt.Sprintf("root:root@tcp(%s:%s)/mysql?charset=utf8mb4&parseTime=True&loc=Local", hostPort.HostIP, hostPort.HostPort)

	return m.Run()
}

//NewDB creates a database connected to the mysql instance in docker.
func NewDB() (*gorm.DB, error) {
	if mysqlDSN == "" {
		return nil, fmt.Errorf("mysql uri not set, Please run RunWithMysqlInDocker in TestMain")
	}

	for {
		if db, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{}); err != nil {
			time.Sleep(4 * time.Second)
		} else {
			return db, err
		}
	}
}
