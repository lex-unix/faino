package app

import (
	"context"
	"fmt"

	"github.com/lex-unix/faino/internal/exec/sshexec"
	"github.com/lex-unix/faino/internal/txman"
)

type RemoteFileContent struct {
	host string
	data []byte
	err  error
}

func ReadRemoteFile(path string, resultsCh chan<- RemoteFileContent) txman.Callback {
	return func(ctx context.Context, client sshexec.Service) error {
		data, err := client.ReadFile(path)
		result := RemoteFileContent{
			host: client.Host(),
			data: data,
			err:  err,
		}
		resultsCh <- result
		if err != nil {
			return fmt.Errorf("host %s: failed to read file %q: %w", client.Host(), path, err)
		}
		return nil
	}
}

func WriteToRemoteFile(path string, data []byte) txman.Callback {
	return func(ctx context.Context, client sshexec.Service) error {
		return client.WriteFile(path, data)
	}
}

func rollbackNoop(_ context.Context, _ sshexec.Service) error { return nil }
