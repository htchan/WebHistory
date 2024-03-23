package updatewebsite

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/rueidis"
)

var cli rueidis.Client

func TestMain(m *testing.M) {
	redisAddr, purge, err := setupContainer()
	if err != nil {
		purge()
		log.Fatalf("fail to setup docker: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	c, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{redisAddr},
	})
	if err != nil {
		purge()
		log.Fatalf("fail to create redis client: %v", err)
	}
	cli = c
	defer cli.Close()

	code := m.Run()

	purge()
	os.Exit(code)
}

func setupContainer() (string, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", func() {}, fmt.Errorf("init docker fail: %w", err)
	}

	containerName := "goworker_test_redis"
	pool.RemoveContainerByName(containerName)

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "redis",
			Tag:        "7",
			Name:       containerName,
		},
		func(hc *docker.HostConfig) {
			hc.AutoRemove = true
		},
	)
	if err != nil {
		return "", func() {}, fmt.Errorf("create resource fail: %w", err)
	}

	purge := func() {
		if resource.Close() != nil {
			fmt.Println("purge error", err)
		}
	}

	return resource.GetHostPort("6379/tcp"), purge, nil
}
