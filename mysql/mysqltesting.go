package mysqltesting

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	unit_test "github.com/Jinrgan/unit-test"
)

const (
	image         = "mysql:5.6"
	containerPort = "3306/tcp"
	pwdEnv        = "MYSQL_ROOT_PASSWORD=root"
	dbEnv         = "MYSQL_DATABASE=test"
)

var mysqlDSN string

const defaultMysqlDSN = "root:123456@tcp(localhost:3306)"

// RunWithMysqlInDocker runs the tests with
// a mysql instance in a docker container.
func RunWithMysqlInDocker(m *testing.M) int {
	return unit_test.RunInDocker(unit_test.DBConfig{
		Image:         image,
		ContainerPort: containerPort,
		Env:           []string{pwdEnv, dbEnv},
		DefaultURI:    defaultMysqlDSN,
		ConnFormatter: func(ip, port string) {
			mysqlDSN = fmt.Sprintf("root:root@tcp(%s:%s)/mysql?charset=utf8mb4&parseTime=True&loc=Local", ip, port)

		},
	}, m)
}

// NewDB creates a database connected to the mysql instance in docker.
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

// NewDefaultDB creates a database connect to localhost:3306
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
