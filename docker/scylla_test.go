package docker_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uol/gotest/docker"
)

// TestScylla - tests the scylla pod
func TestScylla(t *testing.T) {

	pod := "test-scylla-pod"

	docker.Remove(pod)

	ip, err := docker.StartScylla(pod, "", "", 30*time.Second)
	if !assert.NoError(t, err, "error starting scylla pod") {
		return
	}

	defer docker.Remove(pod)

	assert.Regexp(t, regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`), ip, "expected some valid ip")
}
