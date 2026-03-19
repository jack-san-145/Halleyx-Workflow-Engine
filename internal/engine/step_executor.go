package engine

import (
	"bytes"
	"context"
	"fmt"
	"halleyx-workflow-docker/internal/store"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

func ExecuteStep(step *store.Step, ctx map[string]interface{}, execID uuid.UUID) (map[string]interface{}, error) {
	outputs := map[string]interface{}{}
	stepKey := sanitizeKey(step.Name)

	switch step.StepType {
	case "task":
		stdout, stderr, exitCode, err := RunDockerStep(step, ctx, execID)
		outputs["stdout"] = stdout
		outputs["stderr"] = stderr
		outputs["exit_code"] = exitCode

		outputs[fmt.Sprintf("%s_stdout", stepKey)] = stdout
		outputs[fmt.Sprintf("%s_stderr", stepKey)] = stderr
		outputs[fmt.Sprintf("%s_exit_code", stepKey)] = exitCode

		if err != nil {
			return outputs, err
		}
		if exitCode != 0 {
			return outputs, fmt.Errorf("docker step failed with code %d", exitCode)
		}
	case "approval":
		// Pause execution, wait for external callback
		outputs["approval_pending"] = true
	case "notification":
		// Simply log or send webhook
		outputs["notification_sent"] = true
	}
	return outputs, nil
}

func sanitizeKey(input string) string {
	key := strings.ToLower(strings.TrimSpace(input))
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")
	return key
}

// Run Docker container for a step (using Docker SDK instead of `docker` CLI)
func RunDockerStep(step *store.Step, ctx map[string]interface{}, execID uuid.UUID) (string, string, int, error) {
	meta := step.Metadata

	image, _ := meta["image"].(string)
	workdir, _ := meta["workdir"].(string)

	// Replace placeholders
	execIDStr := execID.String()
	if strings.Contains(image, "{{execution_id}}") {
		image = strings.ReplaceAll(image, "{{execution_id}}", execIDStr)
	}
	if strings.Contains(workdir, "{{execution_id}}") {
		workdir = strings.ReplaceAll(workdir, "{{execution_id}}", execIDStr)
	}

	// Support both array and string command definitions
	var cmdArr []string
	if cmdRaw, ok := meta["command"]; ok {
		switch cmd := cmdRaw.(type) {
		case []interface{}:
			for _, c := range cmd {
				cmdArr = append(cmdArr, fmt.Sprintf("%v", c))
			}
		case []string:
			cmdArr = append(cmdArr, cmd...)
		case string:
			cmdArr = []string{cmd}
		}
	}

	if len(cmdArr) == 0 {
		return "", "", -1, fmt.Errorf("no command specified in step metadata")
	}

	volumes := []string{}
	if v, ok := meta["volumes"].([]interface{}); ok {
		for _, vv := range v {
			vol := fmt.Sprintf("%v", vv)
			vol = strings.ReplaceAll(vol, "{{execution_id}}", execIDStr)
			volumes = append(volumes, vol)
		}
	}

	timeout := 60 * time.Second
	if t, ok := meta["timeout"].(float64); ok && t > 0 {
		timeout = time.Duration(t) * time.Second
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", "", -1, err
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	containerConfig := &container.Config{
		Image:        image,
		Cmd:          cmdArr,
		WorkingDir:   workdir,
		AttachStdout: true,
		AttachStderr: true,
	}

	hostConfig := &container.HostConfig{
		Binds:       volumes,
		NetworkMode: "none",
		AutoRemove:  true,
		Resources: container.Resources{
			Memory: 256 * 1024 * 1024,
		},
	}

	resp, err := cli.ContainerCreate(ctxTimeout, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		// If the image is not present locally, attempt to pull it and retry once
		if strings.Contains(err.Error(), "No such image") || strings.Contains(err.Error(), "not found") {
			if pullErr := pullDockerImage(cli, image); pullErr != nil {
				return "", "", -1, fmt.Errorf("failed to pull image %s: %w", image, pullErr)
			}
			resp, err = cli.ContainerCreate(ctxTimeout, containerConfig, hostConfig, nil, nil, "")
			if err != nil {
				return "", "", -1, err
			}
			// proceed with the newly created container
		} else {
			return "", "", -1, err
		}
	}

	if err := cli.ContainerStart(ctxTimeout, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", -1, err
	}

	statusCh, errCh := cli.ContainerWait(ctxTimeout, resp.ID, container.WaitConditionNotRunning)

	var exitCode int
	select {
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	case err := <-errCh:
		return "", "", -1, err
	case <-ctxTimeout.Done():
		return "", "", -1, fmt.Errorf("container timeout")
	}

	logs, err := cli.ContainerLogs(ctxTimeout, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", "", exitCode, err
	}
	defer logs.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(logs)
	out := buf.String()

	return out, "", exitCode, nil
}

func pullDockerImage(cli *client.Client, image string) error {
	// Keep pulling separate from the step timeout (which may be short) so that first-time
	// pulls aren't killed by the container timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Consume the stream to ensure image pull completes.
	_, err = io.Copy(io.Discard, reader)
	return err
}
