package docker_test

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uol/gotest/docker"
)

//
// Has tests for the docker utility functions.
// author: rnojiri
//

func rmPod(pod string) {

	exec.Command("/bin/sh", "-c", "docker rm -f "+pod).Run()
}

func runPod(t *testing.T, pod, image string) bool {

	err := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker run -d --name %s %s", pod, image)).Run()

	return assert.NoError(t, err, "error creating pod")
}

func psPod(t *testing.T, pod string) (string, bool) {

	output, err := exec.Command("/bin/sh", "-c", "docker ps -a -q --filter \"name="+pod+"\"").Output()

	if !assert.NoError(t, err, "error checking pod") {
		return "", false
	}

	return string(output), true
}

func podIP(t *testing.T, pod string) (string, bool) {

	output, err := exec.Command("/bin/sh", "-c", "docker inspect --format='{{ .NetworkSettings.Networks.bridge.IPAddress }}' "+pod).Output()

	if !assert.NoError(t, err, "error inspecting pod") {
		return "", false
	}

	return string(output), true
}

func podExisted(t *testing.T, pod string) bool {

	output, ok := psPod(t, pod)
	if !ok {
		return false
	}

	return assert.Regexp(t, regexp.MustCompile("[0-9a-f]{12}"), output, "expected the pod's hash")
}

// TestRun - tests run command
func TestRun(t *testing.T) {

	pod := "test-run-hello-world"

	rmPod(pod)

	err := docker.Run(pod, "hello-world", "")
	if assert.NoError(t, err, "error not expected") {
		return
	}

	defer rmPod(pod)

	podExisted(t, pod)
}

// TestRemove - tests remove command
func TestRemove(t *testing.T) {

	pod := "test-rm-hello-world"

	rmPod(pod)

	if !runPod(t, pod, "hello-world") {
		return
	}

	if !podExisted(t, pod) {
		return
	}

	err := docker.Remove(pod)
	if !assert.NoError(t, err, "error not expected") {
		return
	}

	output, ok := psPod(t, pod)
	if !assert.True(t, ok, "error on ps") {
		return
	}

	assert.Equal(t, "", output, "expected no output")
}

// TestGetIPs - tests get ips command
func TestGetIPs(t *testing.T) {

	pod := "test-grafana"

	rmPod(pod)

	if !runPod(t, pod, "grafana/grafana") {
		return
	}

	defer rmPod(pod)

	if !podExisted(t, pod) {
		return
	}

	format := ".NetworkSettings.Networks.bridge.IPAddress"

	libIps, err := docker.GetIPs(format, pod)
	if !assert.NoError(t, err, "error not expected") {
		return
	}

	if !assert.Len(t, libIps, 1, "expected only one ip") {
		return
	}

	inspect, ok := podIP(t, pod)
	if !assert.True(t, ok, "error not expected") {
		return
	}

	assert.Contains(t, libIps[0], strings.ReplaceAll(inspect, "\n", ""), "expected same IP")
}

// TestExists - tests exists command
func TestExists(t *testing.T) {

	pod := "test-exists-hello-world"

	rmPod(pod)

	if !runPod(t, pod, "hello-world") {
		return
	}

	if !podExisted(t, pod) {
		return
	}

	defer rmPod(pod)

	exists, err := docker.Exists(pod, docker.Running)
	if !assert.NoError(t, err, "error not expected") {
		return
	}

	if !assert.True(t, exists, "expected pod existed") {
		return
	}

	<-time.After(2 * time.Second)

	exists, err = docker.Exists(pod, docker.Exited)
	if !assert.NoError(t, err, "error not expected") {
		return
	}

	if !assert.True(t, exists, "expected pod existed") {
		return
	}

	rmPod(pod)

	exists, err = docker.Exists(pod, docker.Exited)
	if !assert.NoError(t, err, "error not expected") {
		return
	}

	if !assert.False(t, exists, "expected pod not exists") {
		return
	}
}

// TestWaitUntilListening - wait until the host and port is listening
func TestWaitUntilListening(t *testing.T) {

	address := "localhost:18123"

	go func() {
		<-time.After(1 * time.Second)

		listener, err := net.Listen("tcp", address)
		if assert.NoError(t, err, "expected no error when listening") {
			return
		}

		defer listener.Close()

		c, err := listener.Accept()
		if err != nil {
			if assert.NoError(t, err, "expected no error when accepting a connection") {
				return
			}
		}

		defer c.Close()

		<-time.After(1 * time.Second)
	}()

	connectedMap := docker.WaitUntilListening(3*time.Second, address)

	assert.True(t, connectedMap[address], "expected the address connection")
}
