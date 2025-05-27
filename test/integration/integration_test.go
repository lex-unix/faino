package integration

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	maxRetry = 5
	waitTime = 5 * time.Second
)

func dockerCompose(t *testing.T, composeCmd string) string {
	t.Helper()

	cmd := exec.Command("sh", "-c", fmt.Sprintf("docker compose %s", composeCmd))
	var wg sync.WaitGroup
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	read := func(r io.Reader, w io.Writer) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			t.Log(line)
			fmt.Fprintln(w, line)
		}
		if err := scanner.Err(); err != nil {
			t.Log(err)
		}
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go read(stdoutPipe, &stdout)
	go read(stderrPipe, &stderr)

	wg.Wait()
	if err := cmd.Wait(); err != nil {
		t.Fatalf("command `docker compose %s` failed: %s", composeCmd, err)
	}

	return stdout.String()
}

func deployerExec(t *testing.T, cmd string, workdir string) string {
	return dockerCompose(t, fmt.Sprintf("exec --workdir %s deployer %s", workdir, cmd))
}

func faino(t *testing.T, cmd string) string {
	return deployerExec(t, fmt.Sprintf("/usr/local/bin/faino --debug %s", cmd), "/app")
}

func setup(t *testing.T) {
	t.Log("Setting up docker compose project")
	dockerCompose(t, "up -d --build")
	waitForHealthy(t, maxRetry, waitTime)
	t.Log("Setting up deployer container")
	setupDeployer(t)
	t.Log("Setup successful")

	t.Cleanup(func() {
		t.Log("Cleaning up docker compose project")
		dockerCompose(t, "down -t 1")
	})
}

func setupDeployer(t *testing.T) {
	deployerExec(t, "./setup.sh", "/")
}

func waitForHealthy(t *testing.T, maxRetry int, waitTime time.Duration) {
	for range maxRetry {
		out := dockerCompose(t, "ps -a | tail -n +2 | grep -v '(healthy)' | wc -l")
		out = strings.TrimSpace(out)
		if out == "0" {
			return
		}
		time.Sleep(waitTime)
	}
	t.Fatal("container not healthy")
}

func appResponse() *http.Response {
	res, err := http.Get("http://localhost:3000/")
	if err != nil {
		return nil
	}
	return res
}

func waitForApp(t *testing.T, maxRetry int, waitTime time.Duration) {
	total := 2
	up := 0
	for i := 0; i < maxRetry && up != total; i++ {
		res := appResponse()
		if res != nil && res.StatusCode == http.StatusOK {
			up++
		}
		time.Sleep(waitTime)
	}
	if assert.Equal(t, total, up, "app failed to become ready") == false {
		t.FailNow()
	}
}

func assertAppIsDown(t *testing.T) {
	res := appResponse()
	if assert.Equal(t, http.StatusBadGateway, res.StatusCode, "expected app to be down") == false {
		t.FailNow()
	}
}

func assertAppIsUp(t *testing.T) {
	res := appResponse()
	if assert.Equal(t, http.StatusOK, res.StatusCode, "expected app to be up") == false {
		t.FailNow()
	}
}

func assertAppVersion(t *testing.T, want, got string) {
	if assert.Equal(t, want, got) == false {
		t.FailNow()
	}
}

func getAppVersion() string {
	res := appResponse()
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return string(body)
}
