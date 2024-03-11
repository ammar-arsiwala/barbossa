package main

import (
	"context"
	"log"
	"sync"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func main() {
	var parent_config Config
	err := parent_config.Parse()
	if err != nil {
		log.Fatal(err)
	}

	parent_config.LogCommands = true

	log.SetFlags(log.Lshortfile)
	log.SetPrefix("Barbossa: ")

	context := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	client := Client{
		Client:  cli,
		context: context,
	}

	client.Process(parent_config)
}

type Client struct {
	*client.Client
	context context.Context
}

func (c *Client) Process(config Config) {
	var mtx sync.Mutex
	container := map[string]string{}

	var wg sync.WaitGroup
	for i, service := range config.Basic.Services {
		wg.Add(1)
		go func(i int, service Service) {
			defer wg.Done()

			val := StartContainer(c.context, c.Client, &service)
			mtx.Lock()
			container[service.Name] = val
			mtx.Unlock()
		}(i, service)
	}
	wg.Wait()

	for _, net := range config.Basic.Networks {
		nid, err := CreateNetworkIfNotExists(c.context, c.Client, &net)
		if err != nil {
			log.Panic(err)
		}
		defer c.Client.NetworkRemove(c.context, nid)
		for _, sname := range net.Services {
			endpointSettings := network.EndpointSettings{
				IPAddress: sname.Addr,
			}
			err := c.Client.NetworkConnect(c.context, nid, container[sname.Name], &endpointSettings)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
