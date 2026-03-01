package x_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

func TestInferProjectIDErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid path", func(t *testing.T) {
		t.Parallel()

		_, errE := x.InferGitLabProjectID("/nonexistent/path/does/not/exist")
		require.Error(t, errE)
		assert.ErrorIs(t, errE, x.ErrOpenGitRepository)
	})

	t.Run("no origin remote", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		_, errE := x.InferGitLabProjectID(tempDir)
		require.Error(t, errE)
		assert.ErrorIs(t, errE, x.ErrObtainGitRemote)
	})
}

func TestInferProjectID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		remote    string
		projectID string
	}{
		{"https://gitlab.com/tozd/go/x.git", "tozd/go/x"},
		{"git@gitlab.com:tozd/go/x.git", "tozd/go/x"},
	}

	for k, tt := range tests {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			repository, err := git.PlainInit(tempDir, false)
			require.NoError(t, err)
			workTree, err := repository.Worktree()
			require.NoError(t, err)
			filename := filepath.Join(tempDir, "file.txt")
			author := &object.Signature{
				Name:  "John Doe",
				Email: "john@doe.org",
				When:  time.Now(),
			}
			err = os.WriteFile(filename, []byte("Hello world!"), 0o600)
			require.NoError(t, err)
			_, err = workTree.Add("file.txt")
			require.NoError(t, err)
			_, err = workTree.Commit("Initial commmit.", &git.CommitOptions{
				All:               false,
				AllowEmptyCommits: false,
				Author:            author,
				Committer:         nil,
				Parents:           nil,
				SignKey:           nil,
				Signer:            nil,
				Amend:             false,
			})
			require.NoError(t, err)
			_, err = repository.CreateRemote(&config.RemoteConfig{
				Name:   "origin",
				URLs:   []string{tt.remote},
				Fetch:  nil,
				Mirror: false,
			})
			require.NoError(t, err)
			projectID, errE := x.InferGitLabProjectID(tempDir)
			require.NoError(t, errE, "% -+#.1v", errE)
			assert.Equal(t, tt.projectID, projectID)
		})
	}
}
