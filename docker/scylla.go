package docker

import (
	"errors"
	"time"
)

//
// Commons command execution functions.
// author: rnojiri
//

var (
	// ErrPodNotListening - raised when the pod is not listening
	ErrPodNotListening error = errors.New("pod is not listening")
)

// StartScylla - starts the scylla pod
func StartScylla(pod, extraCommands, networkFormat string, timeout time.Duration) (string, error) {

	err := Run(pod, "scylladb/scylla", extraCommands)
	if err != nil {
		return "", nil
	}

	ips, err := GetIPs(networkFormat, pod)
	if err != nil {
		return "", nil
	}

	if len(ips) == 0 {
		return "", ErrPodNotListening
	}

	connectedMap := WaitUntilListening(timeout, ips[0])

	if _, ok := connectedMap[ips[0]]; !ok {
		return "", ErrPodNotListening
	}

	return ips[0], nil
}
