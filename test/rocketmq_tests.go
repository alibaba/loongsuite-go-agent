// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

const (
	rocketmqModuleName = "rocketmq"
	defaultWaitTimeout = 30 * time.Second
	brokerStartupDelay = 5 * time.Second
)

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("rocketmq_basic-test", rocketmqModuleName, "2.0.0", "", "1.18", "", TestRocketMQBasic),
		NewGeneralTestCase("rocketmq_producer-test", rocketmqModuleName, "2.0.0", "", "1.18", "", TestRocketMQProducer),
		NewGeneralTestCase("rocketmq_consumer-test", rocketmqModuleName, "2.0.0", "", "1.18", "", TestRocketMQConsumer),
	)
}

// TestRocketMQBasic tests basic RocketMQ functionality
func TestRocketMQBasic(t *testing.T, env ...string) {
	containers := initRocketMQContainer(t)
	defer containers.Cleanup(context.Background())

	UseApp("rocketmq/v2.1.2")
	RunGoBuild(t, "go", "build", "test_mq_basic.go", "base.go")

	env = append(env,
		"NAMESRV_ADDR="+containers.NameSrvAddr,
		"BROKER_ADDR="+containers.BrokerAddr,
	)
	RunApp(t, "test_mq_basic", env...)
}

// TestRocketMQProducer tests RocketMQ producer functionality
func TestRocketMQProducer(t *testing.T, env ...string) {
	containers := initRocketMQContainer(t)
	defer containers.Cleanup(context.Background())

	UseApp("rocketmq/v2.1.2")
	RunGoBuild(t, "go", "build", "test_mq_producer.go", "base.go")

	env = append(env,
		"NAMESRV_ADDR="+containers.NameSrvAddr,
		"BROKER_ADDR="+containers.BrokerAddr,
	)
	RunApp(t, "test_mq_producer", env...)
}

// TestRocketMQConsumer tests RocketMQ consumer functionality
func TestRocketMQConsumer(t *testing.T, env ...string) {
	containers := initRocketMQContainer(t)
	defer containers.Cleanup(context.Background())

	UseApp("rocketmq/v2.1.2")
	RunGoBuild(t, "go", "build", "test_mq_consumer.go", "base.go")

	env = append(env,
		"NAMESRV_ADDR="+containers.NameSrvAddr,
		"BROKER_ADDR="+containers.BrokerAddr,
	)
	RunApp(t, "test_mq_consumer", env...)
}

// RocketMQContainers holds references to the RocketMQ containers and their addresses
type RocketMQContainers struct {
	NameSrvContainer testcontainers.Container
	BrokerContainer  testcontainers.Container
	NameSrvAddr      string
	BrokerAddr       string
}

// Cleanup terminates all RocketMQ containers
func (r *RocketMQContainers) Cleanup(ctx context.Context) {
	if r.BrokerContainer != nil {
		_ = r.BrokerContainer.Terminate(ctx)
	}
	if r.NameSrvContainer != nil {
		_ = r.NameSrvContainer.Terminate(ctx)
	}
}

// initRocketMQContainer initializes the RocketMQ test environment with NameServer and Broker
func initRocketMQContainer(t *testing.T) *RocketMQContainers {
	ctx := context.Background()

	// Create a dedicated network for RocketMQ containers
	testNetwork, err := network.New(ctx)
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}

	// Start NameServer container
	nameServerReq := testcontainers.ContainerRequest{
		Image:        "apache/rocketmq:4.9.4",
		ExposedPorts: []string{"9876/tcp"},
		Cmd:          []string{"sh", "-c", "/home/rocketmq/rocketmq-4.9.4/bin/mqnamesrv"},
		WaitingFor:   wait.ForLog("Name Server boot success.").WithStartupTimeout(defaultWaitTimeout),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				"9876/tcp": []nat.PortBinding{{
					HostIP:   "0.0.0.0",
					HostPort: "9876",
				}},
			}
		},
		Networks: []string{testNetwork.Name},
		NetworkAliases: map[string][]string{
			testNetwork.Name: {"rocketmq-nameserver"},
		},
	}

	nameServerC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: nameServerReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start RocketMQ NameServer container: %v", err)
	}

	host := "127.0.0.1"
	nameSrvAddr := host + ":9876"

	// Create broker configuration
	brokerConf := `brokerClusterName=DefaultCluster
brokerName=broker-a
brokerId=0
autoCreateTopicEnable=true
autoCreateSubscriptionGroup=true
deleteWhen=04
fileReservedTime=48
brokerRole=ASYNC_MASTER
flushDiskType=ASYNC_FLUSH
aclEnable=false
diskMaxUsedSpaceRatio=97
brokerIP1=` + host

	// Start Broker container
	brokerReq := testcontainers.ContainerRequest{
		Image:        "apache/rocketmq:4.9.4",
		ExposedPorts: []string{"10911/tcp", "10909/tcp", "10910/tcp"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				"10911/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "10911"}},
				"10909/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "10909"}},
				"10910/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "10910"}},
			}
		},
		Env: map[string]string{
			"NAMESRV_ADDR": "rocketmq-nameserver:9876",
		},
		Networks: []string{testNetwork.Name},
		NetworkAliases: map[string][]string{
			testNetwork.Name: {"rocketmq-broker"},
		},
		Cmd: []string{"sh", "-c",
			"echo '" + brokerConf + "' > /home/rocketmq/rocketmq-4.9.4/conf/broker.conf && " +
				"/home/rocketmq/rocketmq-4.9.4/bin/mqbroker -n rocketmq-nameserver:9876 -c /home/rocketmq/rocketmq-4.9.4/conf/broker.conf"},
		WaitingFor: wait.ForLog("The broker[broker-a").WithStartupTimeout(defaultWaitTimeout * 2),
	}

	brokerC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: brokerReq,
		Started:          true,
	})
	if err != nil {
		nameServerC.Terminate(ctx)
		t.Fatalf("Failed to start Broker container: %v", err)
	}

	brokerAddr := host + ":10911"

	// Wait for Broker to fully initialize
	time.Sleep(brokerStartupDelay)

	return &RocketMQContainers{
		NameSrvContainer: nameServerC,
		BrokerContainer:  brokerC,
		NameSrvAddr:      nameSrvAddr,
		BrokerAddr:       brokerAddr,
	}
}
