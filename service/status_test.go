package service

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	service := Status{
		Name: "asfd",
	}
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	for i := 0; i < 10; i++ {
		resp, err := service.GetStatus(ctx, cli)
		assert.Nil(t, err)
		log.Println(resp)

		time.Sleep(1 * time.Second)
	}
}
