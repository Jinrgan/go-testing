package mongotesting

import (
	"context"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	unit_test "github.com/Jinrgan/go-testing"
)

const (
	image         = "mongo:4.4"
	containerPort = "27017/tcp"
)

var mongoURI string

const defaultMongoURI = "mongodb://localhost:27017"

// RunWithMongoInDocker runs the tests with
// a mongo instance in a docker container.
func RunWithMongoInDocker(m *testing.M) int {
	return unit_test.RunInDocker(unit_test.DBConfig{
		Image:         image,
		ContainerPort: containerPort,
		DefaultURI:    defaultMongoURI,
		ConnFormatter: func(ip, port string) {
			mongoURI = fmt.Sprintf("mongodb://%s:%s", ip, port)
		},
	}, m)
}

// NewClient creates a client connected to the mongo instance in docker.
func NewClient(c context.Context) (*mongo.Client, error) {
	if mongoURI == "" {
		return nil, fmt.Errorf("mongo uri not set, Please run RunWithMongoInDocker in TestMain")
	}

	return mongo.Connect(c, options.Client().ApplyURI(mongoURI))
}

// NewDefaultClient creates a client connected to localhost:27017
func NewDefaultClient(c context.Context) (*mongo.Client, error) {
	return mongo.Connect(c, options.Client().ApplyURI(defaultMongoURI))
}
