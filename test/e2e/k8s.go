//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// K8sHelper provides utility methods for interacting with a Kubernetes cluster via kubectl.
type K8sHelper struct {
	Namespace string
}

// NewK8sHelper returns a helper for executing kubectl commands in a namespace.
func NewK8sHelper(namespace string) *K8sHelper {
	return &K8sHelper{
		Namespace: namespace,
	}
}

// Exec runs a command inside a specific pod container and returns stdout.
func (k *K8sHelper) Exec(ctx context.Context, pod string, args ...string) (string, error) {
	cmdArgs := []string{"exec", "-n", k.Namespace, pod, "--"}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.CommandContext(ctx, "kubectl", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("kubectl exec failed: %w, stderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// Logs returns the stdout/stderr logs of a pod.
func (k *K8sHelper) Logs(ctx context.Context, pod string) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "logs", "-n", k.Namespace, pod)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("kubectl logs failed: %w, stderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// DescribePod returns the output of kubectl describe pod.
func (k *K8sHelper) DescribePod(ctx context.Context, pod string) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "describe", "pod", "-n", k.Namespace, pod)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// GetPodsByLabel returns a list of pod names matching the given label selector.
func (k *K8sHelper) GetPodsByLabel(ctx context.Context, labelSelector string) ([]string, error) {
	cmd := exec.CommandContext(
		ctx,
		"kubectl",
		"get",
		"pods",
		"-n",
		k.Namespace,
		"-l",
		labelSelector,
		"-o",
		"jsonpath={.items[*].metadata.name}",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("kubectl get pods failed: %w, output: %s", err, string(out))
	}
	names := strings.Fields(string(out))
	return names, nil
}

// WaitForPodRunning waits for a pod to transition to Running phase.
func (k *K8sHelper) WaitForPodRunning(ctx context.Context, pod string, timeout time.Duration) error {
	timeoutStr := fmt.Sprintf("%ds", int(timeout.Seconds()))
	cmd := exec.CommandContext(
		ctx,
		"kubectl",
		"wait",
		"-n",
		k.Namespace,
		"--for=jsonpath={.status.phase}=Running",
		"pod/"+pod,
		"--timeout="+timeoutStr,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl wait failed: %w, output: %s", err, string(out))
	}
	return nil
}

// StartPortForward spawns a kubectl port-forward process in the background and returns a cleanup func.
func (k *K8sHelper) StartPortForward(ctx context.Context, pod string, localPort, remotePort int) (func(), error) {
	portMapping := fmt.Sprintf("%d:%d", localPort, remotePort)
	portForwardCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(portForwardCtx, "kubectl", "port-forward", "-n", k.Namespace, "pod/"+pod, portMapping)

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start port-forward: %w", err)
	}

	stopFunc := func() {
		cancel()
		_ = cmd.Wait()
	}

	// Wait for port to open
	addr := fmt.Sprintf("localhost:%d", localPort)
	deadline := time.Now().Add(10 * time.Second)
	dialer := &net.Dialer{Timeout: 200 * time.Millisecond}
	for time.Now().Before(deadline) {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return stopFunc, nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Return stopFunc even if dial timed out, so caller can clean up
	return stopFunc, nil
}
