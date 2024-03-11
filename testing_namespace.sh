#!/bin/bash

set -xe

ip netns add red
ip netns add blue

function cleanup() {
	ip netns del red
	ip netns del blue
}

trap cleanup EXIT

ip link add veth-red type veth peer name veth-blue
ip link set veth-red netns red
ip link set veth-blue netns blue

ip -n red addr add 10.0.0.1 dev veth-red
ip -n blue addr add 10.0.0.2 dev veth-blue

ip -n red link set veth-red up
ip -n blue link set veth-blue up

ip netns exec red ip route add default via 10.0.0.1 dev veth-red
ip netns exec blue ip route add default via 10.0.0.2 dev veth-blue

ip netns exec red ping -c 5 10.0.0.2
ip netns exec red ping -c 5 127.0.0.1
