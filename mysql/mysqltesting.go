package mysqltesting

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

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

const defaultMysqlDSN = "root:123456@tcp(localhost:3306)"

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

//NewDefaultDB creates a database connect to localhost:3306
func NewDefaultDB(db string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(defaultMysqlDSN+"/"+db), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 启用单数命名
		},
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Millisecond, // 慢查询阈值
				Colorful:      true,
				LogLevel:      logger.Info,
			}),
	})
}
