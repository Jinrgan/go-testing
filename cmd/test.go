package cmd

import (
	"context"
	"os"
	"testing"

	mongotesting "github.com/Jinrgan/unit-test/mongo"
)

func TestTest(t *testing.T) {
	ctx := context.Background()
	clt, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Errorf("cannot create mongo client: %v", err)
	}
	_ = clt.Database("matchmaking")

}

func TestMain(m *testing.M) {
	// 运行 MongoDB 在 Docker 中，并处理可能的错误
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
