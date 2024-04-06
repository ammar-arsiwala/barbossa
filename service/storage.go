package service

import (
	"sync"

	"github.com/docker/docker/api/types"
)

type EndpointInformationStorage interface {
	Store(networkID, containerID string, resource types.EndpointResource)
	Get(networkID, containerID string) (types.EndpointResource, bool)
}

type networkAndContainer struct {
	N, C string
}

type NaiveEndpointStorage struct {
	m *sync.Mutex
	s map[networkAndContainer]types.EndpointResource
}

func (n *NaiveEndpointStorage) Store(networkID, containerID string, resource types.EndpointResource) {
	key := networkAndContainer{
		N: networkID,
		C: containerID,
	}

	n.s[key] = resource
}

func (n *NaiveEndpointStorage) Get(networkID, containerID string) (types.EndpointResource, bool) {
	key := networkAndContainer{
		N: networkID,
		C: containerID,
	}
	e, f := n.s[key]
	return e, f
}

func NewNaiveEndpointStorage() EndpointInformationStorage {
	return &NaiveEndpointStorage{
		m: &sync.Mutex{},
		s: make(map[networkAndContainer]types.EndpointResource),
	}
}

var defaultEndpointStorage = NewNaiveEndpointStorage()
