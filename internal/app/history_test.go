package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/lex-unix/faino/internal/exec/sshexec"
	"github.com/lex-unix/faino/internal/txman"
	"github.com/stretchr/testify/assert"
)

var testHistoryData = []byte(`[
	{"version": "3", "timestamp": "2025-05-01T10:00:00.000Z"},
	{"version": "1", "timestamp": "2025-02-01T10:00:00.000Z"},
	{"version": "2", "timestamp": "2025-03-01T10:00:00.000Z"}
]`)

type StubSSH struct {
	out bytes.Buffer
}

func (stub *StubSSH) Run(ctx context.Context, cmd string, options ...sshexec.SessionOption) error {
	return nil
}

func (stub *StubSSH) ReadFile(path string) ([]byte, error) {
	return nil, errors.New("stub Readfile()")
}

func (stub *StubSSH) WriteFile(path string, data []byte) error {
	stub.out.Reset()
	_, err := stub.out.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (stub StubSSH) Host() string {
	return "test@host"
}

type StubTxman struct {
	ssh *StubSSH
}

func (stub *StubTxman) Execute(ctx context.Context, callback txman.Callback) error {
	return nil
}

func (stub *StubTxman) SetPrimaryHost(_ string) error {
	return nil
}

func TestHistorySort(t *testing.T) {
	app := &App{}
	err := app.loadHistory(testHistoryData)
	assert.NoError(t, err)

	app.sortHistory()

	var historyData []HistoryEntry
	err = json.Unmarshal(testHistoryData, &historyData)
	assert.NoError(t, err)

	assert.Len(t, app.history, len(historyData))
	assert.Equal(t, app.history[0].Version, "3")
	assert.Equal(t, app.history[1].Version, "2")
	assert.Equal(t, app.history[2].Version, "1")
}

func TestHistoryLatestVersion(t *testing.T) {
	app := &App{}
	err := app.loadHistory(testHistoryData)
	assert.NoError(t, err)

	got := app.LatestVersion()
	expected := "3"
	assert.Equal(t, expected, got)
}

// func TestHistoryAppend(t *testing.T) {
// 	ssh := &StubSSH{}
// 	txmanager := &StubTxman{ssh: ssh}
// 	app := &App{txmanager: txmanager}
// 	err := app.loadHistory(testHistoryData)
// 	assert.NoError(t, err)
//
// 	initialLen := len(app.history)
// 	appendedVersion := "4"
//
// 	err = app.AppendVersion(appendedVersion)(context.TODO(), txmanager.ssh)
// 	assert.NoError(t, err)
//
// 	actualWrittenBytes := ssh.out.String()
// 	expectedJSONBytes, err := json.Marshal(app.history)
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, appendedVersion, app.history[initialLen].Version)
// 	assert.Len(t, app.history, initialLen+1)
// 	assert.JSONEq(t, string(expectedJSONBytes), actualWrittenBytes)
// }
