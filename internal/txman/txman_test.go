package txman

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/lex-unix/faino/internal/exec/sshexec"
	"github.com/stretchr/testify/assert"
)

func TestBeginTransactin(t *testing.T) {
	t.Run("runs every command on each host", func(t *testing.T) {
		var mu sync.Mutex
		calls := make(map[string][]string)
		sshClient1 := NewMockSSHLikeService("host1")
		sshClient1.RunFunc = func(ctx context.Context, cmd string, options ...sshexec.SessionOption) error {
			mu.Lock()
			defer mu.Unlock()
			calls[sshClient1.hostName] = append(calls[sshClient1.hostName], cmd)
			return nil
		}

		sshClient2 := NewMockSSHLikeService("host1")
		sshClient2.RunFunc = func(ctx context.Context, cmd string, options ...sshexec.SessionOption) error {
			mu.Lock()
			defer mu.Unlock()
			calls[sshClient2.hostName] = append(calls[sshClient2.hostName], cmd)
			return nil
		}

		m := New(sshClient1, sshClient2)

		ctx := context.Background()
		_, err := m.BeginTransaction(ctx, func(ctx context.Context, tx Transaction) error {
			if err := tx.Run(ctx, "command 1", "rollback command 1"); err != nil {
				return err
			}
			if err := tx.Run(ctx, "command 2", "rollback command 2"); err != nil {
				return err
			}
			if err := tx.Run(ctx, "command 3", "rollback command 3"); err != nil {
				return err
			}
			return nil
		})

		calls1 := calls[sshClient1.hostName]
		calls2 := calls[sshClient2.hostName]

		assert.NoError(t, err)
		assert.NotEmpty(t, calls1)
		assert.NotEmpty(t, calls2)
		assert.ElementsMatch(t, calls1, calls2)
		assert.Equal(t, "command 1", calls1[0])
		assert.Equal(t, "command 2", calls1[1])
		assert.Equal(t, "command 3", calls1[2])
		assert.Equal(t, "command 1", calls2[0])
		assert.Equal(t, "command 2", calls2[1])
		assert.Equal(t, "command 3", calls2[2])
	})

	t.Run("returns errors if forward pass and rollback pass fail", func(t *testing.T) {
		// waitCh is used to sync command execution progress between ssh clients
		waitCh := make(chan struct{})
		sshClient1 := NewMockSSHLikeService("host1")
		sshClient1.RunFunc = func(ctx context.Context, cmd string, options ...sshexec.SessionOption) error {
			if cmd == "command 2" {
				select {
				// wait for sshClient2 to finish executing 'command 1'
				case <-waitCh:
					return fmt.Errorf("host %s failed on command: %s", sshClient1.hostName, cmd)
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		}

		sshClient2 := NewMockSSHLikeService("host2")
		sshClient2.RunFunc = func(ctx context.Context, cmd string, options ...sshexec.SessionOption) error {
			if cmd == "command 1" {
				// signal sshClient1 that 'command 1' has completed
				close(waitCh)
				return nil
			}
			if cmd == "rollback command 1" {
				return fmt.Errorf("host %s failed rollback command: %s", sshClient2.hostName, cmd)
			}
			return nil
		}

		m := New(sshClient1, sshClient2)

		ctx := context.Background()
		rollback, err := m.BeginTransaction(ctx, func(ctx context.Context, tx Transaction) error {
			if err := tx.Run(ctx, "command 1", "rollback command 1"); err != nil {
				return err
			}
			if err := tx.Run(ctx, "command 2", "rollback command 2"); err != nil {
				return err
			}
			return nil
		})

		assert.Error(t, err)

		ctx = context.Background()
		err = rollback(ctx)
		assert.Error(t, err)
	})

	t.Run("executes rollback functions in correct order", func(t *testing.T) {
		rollbackCmds := make([]string, 0)
		sshClient := NewMockSSHLikeService("host1")
		sshClient.RunFunc = func(ctx context.Context, cmd string, options ...sshexec.SessionOption) error {
			if cmd == "command 3" {
				return errors.New("command failed")
			}
			if strings.HasPrefix(cmd, "rollback") {
				rollbackCmds = append(rollbackCmds, cmd)
			}
			return nil
		}

		m := New(sshClient)

		ctx := context.Background()
		rollback, err := m.BeginTransaction(ctx, func(ctx context.Context, tx Transaction) error {
			if err := tx.Run(ctx, "command 1", "rollback 1"); err != nil {
				return err
			}
			if err := tx.Run(ctx, "command 2", "rollback 2"); err != nil {
				return err
			}
			if err := tx.Run(ctx, "command 3", "rollback 3"); err != nil {
				return err
			}
			return nil
		})

		assert.Error(t, err)
		err = rollback(context.Background())
		assert.NoError(t, err)
		assert.Len(t, rollbackCmds, 2)
		assert.Equal(t, "rollback 2", rollbackCmds[0])
		assert.Equal(t, "rollback 1", rollbackCmds[1])
	})
}
