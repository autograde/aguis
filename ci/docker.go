package ci

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Docker is an implementation of the CI interface using Docker.
type Docker struct {
	Endpoint string
	Version  string
}

var (
	containerTimeout = time.Duration(10 * time.Minute)
	stopTimeout      = time.Duration(5 * time.Second)
)

// Run implements the CI interface. This method blocks until the job has been
// completed or an error occurs, e.g., the context times out.
func (d *Docker) Run(ctx context.Context, job *Job, user string, timeout time.Duration) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}

	if err := pullImage(ctx, cli, job.Image); err != nil {
		return "", err
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: job.Image,
		Cmd:   []string{"/bin/sh", "-c", strings.Join(job.Commands, "\n")},
	}, nil, nil, user)
	if err != nil {
		return "", err
	}

	if csErr := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); csErr != nil {
		return "", csErr
	}

	// will wait until the container stops
	waitc, errc := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	if timeout < 1 {
		timeout = containerTimeout
	}

	select {
	case wErr := <-errc:
		fmt.Println("wErr: ", wErr.Error())
		return "", wErr
		// if the container still running after predefined time interval, force kill it
	case <-time.After(timeout):
		fmt.Println("Timeout done, stopping container: ", user)
		cli.ContainerKill(ctx, resp.ID, "SIGTERM")
		return fmt.Sprintf("Container timed out after %d minutes", timeout), nil
	case <-waitc:
	}

	r, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, ioutil.Discard, r); err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func pullImage(ctx context.Context, cli *client.Client, image string) error {
	progress, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer progress.Close()

	_, err = io.Copy(ioutil.Discard, progress)
	return err
}
