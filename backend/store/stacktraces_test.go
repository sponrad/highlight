package store

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/smithy-go/ptr"
	github2 "github.com/google/go-github/v50/github"
	"github.com/highlight-run/highlight/backend/model"
	privateModel "github.com/highlight-run/highlight/backend/private-graph/graph/model"
	"github.com/stretchr/testify/assert"
)

func TestGitHubFilePath(t *testing.T) {
	defer teardown(t)
	var tests = []struct {
		FileName     string
		BuildPrefix  *string
		GitHubPrefix *string
		Expected     string
	}{
		{
			FileName:     "/build/backend/util/tracer.go",
			BuildPrefix:  ptr.String("/build"),
			GitHubPrefix: ptr.String("/src"),
			Expected:     "/src/backend/util/tracer.go",
		},
		{
			FileName:    "/build/backend/util/tracer.go",
			BuildPrefix: ptr.String("/build"),
			Expected:    "/backend/util/tracer.go",
		},
		{
			FileName:     "/build/backend/util/tracer.go",
			GitHubPrefix: ptr.String("/src"),
			Expected:     "/src/build/backend/util/tracer.go",
		},
		{
			FileName: "/build/backend/util/tracer.go",
			Expected: "/build/backend/util/tracer.go",
		},
		{
			FileName:     "/build/backend/util/tracer.go",
			BuildPrefix:  ptr.String("/foo"),
			GitHubPrefix: ptr.String("/bar"),
			Expected:     "/build/backend/util/tracer.go",
		},
		{
			FileName:    "/build/backend/util/tracer.go",
			BuildPrefix: ptr.String("/foo"),
			Expected:    "/build/backend/util/tracer.go",
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		newPath := store.GitHubFilePath(ctx, tt.FileName, tt.BuildPrefix, tt.GitHubPrefix)
		assert.Equal(t, tt.Expected, newPath)
	}
}

func TestExpandedStackTrace(t *testing.T) {
	defer teardown(t)
	lines := []string{"b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "s", "v", "w", "x", "y", "z"}
	var tests = []struct {
		LineNumber      int
		ExpectedContent string
		ExpectedBefore  string
		ExpectedAfter   string
		ExpectedError   bool
	}{
		{
			LineNumber:      7,
			ExpectedContent: "h",
			ExpectedBefore:  "c\nd\ne\nf\ng",
			ExpectedAfter:   "i\nj\nk\nl\nm",
			ExpectedError:   false,
		},
		{
			LineNumber:      2,
			ExpectedContent: "c",
			ExpectedBefore:  "b",
			ExpectedAfter:   "d\ne\nf\ng\nh",
			ExpectedError:   false,
		},
		{
			LineNumber:      21,
			ExpectedContent: "z",
			ExpectedBefore:  "s\nv\nw\nx\ny",
			ExpectedAfter:   "",
			ExpectedError:   false,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		content, before, after, err := store.ExpandedStackTrace(ctx, lines, tt.LineNumber)
		assert.Equal(t, tt.ExpectedContent, *content)
		assert.Equal(t, tt.ExpectedBefore, *before)
		assert.Equal(t, tt.ExpectedAfter, *after)
		if tt.ExpectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestFetchFileFromGitHub(t *testing.T) {
	defer teardown(t)
	var tests = []struct {
		Trace           *privateModel.ErrorTrace
		Service         *model.Service
		FileName        string
		ServiceVersion  string
		ExpectedContent *string
		ExpectedError   bool
	}{
		{
			Trace: &privateModel.ErrorTrace{
				FileName:     ptr.String("/build/file.js"),
				LineNumber:   ptr.Int(634),
				ColumnNumber: ptr.Int(4),
				FunctionName: ptr.String(""),
			},
			Service: &model.Service{
				GithubRepoPath: ptr.String("highlight/highlight"),
			},
			FileName:        "/file.js",
			ServiceVersion:  "1234567890",
			ExpectedContent: ptr.String("Y29uc29sZS5sb2coJ2hlbGxvIHdvcmxkJyk="),
			ExpectedError:   false,
		},
		{
			Trace: &privateModel.ErrorTrace{
				FileName:     ptr.String("/build/error.js"),
				LineNumber:   ptr.Int(634),
				ColumnNumber: ptr.Int(4),
				FunctionName: ptr.String(""),
			},
			Service: &model.Service{
				GithubRepoPath: ptr.String("highlight/highlight"),
			},
			FileName:        "/error.js",
			ServiceVersion:  "1234567890",
			ExpectedContent: nil,
			ExpectedError:   true,
		},
		{
			Trace: &privateModel.ErrorTrace{
				FileName:     ptr.String("/build/blob-file.js"),
				LineNumber:   ptr.Int(634),
				ColumnNumber: ptr.Int(4),
				FunctionName: ptr.String(""),
			},
			Service: &model.Service{
				GithubRepoPath: ptr.String("highlight/highlight"),
			},
			FileName:        "/blob-file.js",
			ServiceVersion:  "1234567890",
			ExpectedContent: ptr.String("Y29uc29sZS5sb2coJ2hlbGxvIGJpZyB3b3JsZCcp"),
			ExpectedError:   false,
		},
		{
			Trace: &privateModel.ErrorTrace{
				FileName:     ptr.String("/build/blob-error.js"),
				LineNumber:   ptr.Int(634),
				ColumnNumber: ptr.Int(4),
				FunctionName: ptr.String(""),
			},
			Service: &model.Service{
				GithubRepoPath: ptr.String("highlight/highlight"),
			},
			FileName:        "/blob-error.js",
			ServiceVersion:  "1234567890",
			ExpectedContent: nil,
			ExpectedError:   true,
		},
	}

	ctx := context.Background()
	githubClientMock := MockGithubClient{}

	for _, tt := range tests {
		content, err := store.FetchFileFromGitHub(ctx, tt.Trace, tt.Service, tt.FileName, tt.ServiceVersion, &githubClientMock)
		if tt.ExpectedError {
			assert.Nil(t, content)
			assert.Error(t, err)
		} else {
			assert.Equal(t, *tt.ExpectedContent, *content)
			assert.NoError(t, err)
		}
	}
}

func TestGitHubGitSHA(t *testing.T) {
	defer teardown(t)
	var tests = []struct {
		GitHubRepoPath string
		ServiceVersion string
		ExpectedSHA    *string
		ExpectedError  bool
	}{
		{
			GitHubRepoPath: "highlight/highlight",
			ServiceVersion: "1234567890",
			ExpectedSHA:    ptr.String("1234567890"),
			ExpectedError:  false,
		},
		{
			GitHubRepoPath: "highlight/error",
			ServiceVersion: "invalid-sha",
			ExpectedSHA:    nil,
			ExpectedError:  true,
		},
		{
			GitHubRepoPath: "highlight/found",
			ServiceVersion: "invalid-sha",
			ExpectedSHA:    ptr.String("0987654321"),
			ExpectedError:  false,
		},
	}

	ctx := context.Background()
	githubClientMock := MockGithubClient{}

	for _, tt := range tests {
		sha, err := store.GitHubGitSHA(ctx, tt.GitHubRepoPath, tt.ServiceVersion, &githubClientMock)
		if tt.ExpectedError {
			assert.Nil(t, sha)
			assert.Error(t, err)
		} else {
			assert.Equal(t, *tt.ExpectedSHA, *sha)
			assert.NoError(t, err)
		}
	}
}

func TestEnhanceTraceWithGitHub(t *testing.T) {
	defer teardown(t)
	// test no file name or line number, test matches ignore config, found
	var tests = []struct {
		Trace              *privateModel.ErrorTrace
		Service            *model.Service
		ServiceVersion     string
		IgnoredFiles       []string
		ExpectedErrorTrace *privateModel.ErrorTrace
		ExpectedError      bool
	}{
		{
			Trace: &privateModel.ErrorTrace{
				FileName:     ptr.String("/file.js"),
				LineNumber:   ptr.Int(1),
				ColumnNumber: ptr.Int(4),
				FunctionName: ptr.String(""),
			},
			Service: &model.Service{
				GithubRepoPath: ptr.String("highlight/highlight"),
			},
			ServiceVersion: "1234567890",
			IgnoredFiles:   []string{".*/node_modules/.*", ".*/go/pkg/mod/.*"},
			ExpectedErrorTrace: &privateModel.ErrorTrace{
				LineContent: ptr.String("console.log('hello world')"),
			},
			ExpectedError: false,
		},
		{
			Trace: &privateModel.ErrorTrace{
				FileName:     ptr.String("/file.js"),
				LineNumber:   nil,
				ColumnNumber: ptr.Int(4),
				FunctionName: ptr.String(""),
			},
			Service: &model.Service{
				GithubRepoPath: ptr.String("highlight/highlight"),
			},
			ServiceVersion:     "1234567890",
			IgnoredFiles:       []string{".*/node_modules/.*", ".*/go/pkg/mod/.*"},
			ExpectedErrorTrace: nil,
			ExpectedError:      true,
		},
		{
			Trace: &privateModel.ErrorTrace{
				FileName:     ptr.String("/backend/go/pkg/mod/file.js"),
				LineNumber:   ptr.Int(1),
				ColumnNumber: ptr.Int(4),
				FunctionName: ptr.String(""),
			},
			Service: &model.Service{
				GithubRepoPath: ptr.String("highlight/highlight"),
			},
			ServiceVersion:     "1234567890",
			IgnoredFiles:       []string{".*/node_modules/.*", ".*/go/pkg/mod/.*"},
			ExpectedErrorTrace: nil,
			ExpectedError:      false,
		},
	}

	ctx := context.Background()
	githubClientMock := MockGithubClient{}

	for _, tt := range tests {
		errorTrace, err := store.EnhanceTraceWithGitHub(ctx, tt.Trace, tt.Service, tt.ServiceVersion, tt.IgnoredFiles, &githubClientMock)

		if tt.ExpectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		if tt.ExpectedErrorTrace != nil {
			// builds new trace off initial trace
			assert.Equal(t, *tt.Trace.FileName, *errorTrace.FileName)
			assert.Equal(t, *tt.ExpectedErrorTrace.LineContent, *errorTrace.LineContent)
		} else {
			// returns intial trace
			assert.Equal(t, *tt.Trace.FileName, *errorTrace.FileName)
			assert.Nil(t, errorTrace.LineContent)
		}
	}
}

type MockGithubClient struct{}

func (c *MockGithubClient) GetRepoContent(ctx context.Context, githubPath string, path string, version string) (fileContent *github2.RepositoryContent, directoryContent []*github2.RepositoryContent, resp *github2.Response, err error) {
	if path == "/error.js" {
		return nil, nil, nil, errors.New("repo error")
	}
	if path == "/file.js" {
		fileContent := github2.RepositoryContent{
			// base64 for console.log('hello world')
			Content: ptr.String("Y29uc29sZS5sb2coJ2hlbGxvIHdvcmxkJyk="),
		}
		return &fileContent, nil, nil, nil
	}
	if path == "/blob-error.js" {
		emptyContent := github2.RepositoryContent{
			Content: ptr.String(""),
			SHA:     ptr.String("blob-error"),
		}
		return &emptyContent, nil, nil, nil
	}
	if path == "/blob-file.js" {
		emptyContent := github2.RepositoryContent{
			Content: ptr.String(""),
			SHA:     ptr.String("blob-file"),
		}
		return &emptyContent, nil, nil, nil
	}
	return nil, nil, nil, nil
}

func (c *MockGithubClient) GetRepoBlob(ctx context.Context, githubPath string, blobSHA string) (*github2.Blob, *github2.Response, error) {
	if blobSHA == "blob-error" {
		return nil, nil, errors.New("blob error")
	}
	if blobSHA == "blob-file" {
		blobContent := github2.Blob{
			// base64 for console.log('hello big world')
			Content: ptr.String("Y29uc29sZS5sb2coJ2hlbGxvIGJpZyB3b3JsZCcp"),
		}
		return &blobContent, nil, nil
	}

	return nil, nil, nil
}

func (c *MockGithubClient) GetLatestCommitHash(ctx context.Context, githubPath string) (string, *github2.Response, error) {
	if githubPath == "highlight/error" {
		return "", nil, errors.New("error")
	}
	if githubPath == "highlight/found" {
		return "0987654321", nil, nil
	}
	return "", nil, nil
}

// other methods not used in this test but needed for interface
func (c *MockGithubClient) CreateIssue(ctx context.Context, repo string, issueRequest *github2.IssueRequest) (*github2.Issue, error) {
	return nil, nil
}
func (c *MockGithubClient) ListLabels(ctx context.Context, repo string) ([]*github2.Label, error) {
	return nil, nil
}
func (c *MockGithubClient) ListRepos(ctx context.Context) ([]*github2.Repository, error) {
	return nil, nil
}
func (c *MockGithubClient) DeleteInstallation(ctx context.Context, installation string) error {
	return nil
}