package docker

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

//
// Commons command execution functions.
// author: rnojiri
//

var (
	// ErrPodHashNotFound - raised when the output is incompatible with the hash pattern
	ErrPodHashNotFound error = errors.New("pod hash pattern was not found")

	regexpPodHashPattern         *regexp.Regexp = regexp.MustCompile("[a-f0-9]{64}")
	regexpDirtChars              *regexp.Regexp = regexp.MustCompile(`["'\r]+`)
	regexpDirtCharsAndLineBreaks *regexp.Regexp = regexp.MustCompile(`["'\r\n]+`)
)

// PodStatus - the pod status to be filtered
type PodStatus string

const (
	// Restarting - pod status
	Restarting PodStatus = "restarting"

	// Running - pod status
	Running PodStatus = "running"

	// Removing - pod status
	Removing PodStatus = "removing"

	// Paused - pod status
	Paused PodStatus = "paused"

	// Exited - pod status
	Exited PodStatus = "exited"

	// Dead - pod status
	Dead PodStatus = "dead"
)

// createDockerCommand - creates a docker command to run or output
func createDockerCommand(cmd string) *exec.Cmd {

	return exec.Command("/bin/sh", "-c", fmt.Sprintf("docker %s", cmd))
}

// Run - runs a pod
func Run(name, image, extra string) error {

	output, err := createDockerCommand(fmt.Sprintf("run --name %s %s -d %s", name, extra, image)).Output()
	if err != nil {
		return err
	}

	podHash := strings.Split(string(output), "\n")[0]

	if !regexpPodHashPattern.MatchString(podHash) {
		return ErrPodHashNotFound
	}

	return nil
}

// Remove - removes a pod
func Remove(pod string) error {

	return createDockerCommand(fmt.Sprintf("rm -f %s", pod)).Run()
}

// GetIPs - return the pod's ips
func GetIPs(format string, pod ...string) ([]string, error) {

	if len(format) == 0 {
		format = ".NetworkSettings.Networks.bridge.IPAddress"
	}

	output, err := createDockerCommand(fmt.Sprintf("inspect --format='{{ %s }}' %s", format, strings.Join(pod, " "))).Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(regexpDirtChars.ReplaceAllString(string(output), ""), "\n")

	return lines[0 : len(lines)-1], nil
}

// Exists - check if a pod exists
func Exists(pod string, status PodStatus) (bool, error) {

	output, err := createDockerCommand(fmt.Sprintf(`ps -a -q --filter "name=%s" --filter "status=%s" --format "{{.Names}}"`, pod, status)).Output()
	if err != nil {
		return false, err
	}

	return regexpDirtCharsAndLineBreaks.ReplaceAllString(string(output), "") == pod, nil
}

// WaitUntilListening - wait some pod(s) to be listening
func WaitUntilListening(timeout time.Duration, address ...string) map[string]bool {

	start := time.Now()
	connected := map[string]bool{}
	wg := sync.WaitGroup{}
	wg.Add(len(address))

	testConn := func(i int) {

		for {

			if time.Now().Sub(start).Seconds() > timeout.Seconds() {
				connected[address[i]] = false
				break
			}

			<-time.After(1 * time.Second)
			conn, err := net.DialTimeout("tcp", address[i], 1*time.Second)
			if err != nil {
				continue
			}

			if conn != nil {
				defer conn.Close()
				connected[address[i]] = true
				break
			}
		}

		wg.Done()
	}

	for i := 0; i < len(address); i++ {

		testConn(i)
	}

	wg.Wait()

	return connected
}
