package integration

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	setup(t)

	t.Log("running `faino deploy`")
	faino(t, "deploy")

	waitForApp(t, maxRetry, waitTime)

	assertAppIsUp(t)

	t.Log("running `faino app info`")
	info := faino(t, "app show")

	assert.Regexp(t, regexp.MustCompile("Host vm1:"), info)
	assert.Regexp(t, regexp.MustCompile("Host vm2:"), info)

	t.Log("running `faino app stop`")
	faino(t, "app stop")

	assertAppIsDown(t)

	t.Log("running `faino app start`")
	faino(t, "app start")

	waitForApp(t, maxRetry, waitTime)

	assertAppIsUp(t)
}

// func TestRollback(t *testing.T) {
// 	setup(t)
//
// 	t.Log("running `faino deploy`")
// 	faino(t, "deploy")
//
// 	waitForApp(t, maxRetry, waitTime)
//
// 	assertAppIsUp(t)
//
// 	want := getAppVersion()
//
// 	deployerExec(t, "git commit --amend -am \"second commit\"", "/app")
// 	deployerExec(t, "sh -c 'git rev-parse --short HEAD > version.txt' ", "/app")
//
// 	t.Log("running `faino deploy`")
// 	faino(t, "deploy")
//
// 	waitForApp(t, maxRetry, waitTime)
//
// 	assertAppIsUp(t)
//
// 	t.Logf("running `faino rollback %s`", want[:7])
// 	faino(t, fmt.Sprintf("rollback %s", want[:7]))
//
// 	waitForApp(t, maxRetry, waitTime)
//
// 	assertAppIsUp(t)
//
// 	assertAppVersion(t, want, getAppVersion())
// }
