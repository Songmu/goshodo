package goshodo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"
)

const userAgentBase = "Songmu-goshodo/%s (+https://github.com/Songmu/goshodo)"

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

type client struct {
	c           doer
	root, token string
}

func newClient() (*client, error) {
	root, ok := os.LookupEnv("SHODO_API_ROOT")
	if !ok {
		return nil, fmt.Errorf("no SHODO api root")
	}
	token, ok := os.LookupEnv("SHODO_API_TOKEN")
	if !ok {
		return nil, fmt.Errorf("no SHODO api token")
	}
	return &client{
		c:     http.DefaultClient,
		root:  root,
		token: token,
	}, nil
}

func (c *client) newRequest(ctx context.Context, method, p string, body io.Reader) *http.Request {
	p, _ = url.JoinPath(c.root, p)
	req, _ := http.NewRequestWithContext(ctx, method, p, body)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", fmt.Sprintf(userAgentBase, version))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// https://github.com/zenproducts/developers.shodo.ink/blob/master/docs/api.md
func (c *client) CreateLint(ctx context.Context, body string) (string, error) {
	payload := map[string]string{"body": body}
	j, _ := json.Marshal(payload)
	b := bytes.NewReader(j)
	resp, err := c.c.Do(c.newRequest(ctx, http.MethodPost, "lint/", b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned response code: %d", resp.StatusCode)
	}
	var r struct {
		LintID string `json:"lint_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.LintID, nil
}

type LintPosition struct {
	Ch   int `json:"ch"`
	Line int `json:"line"`
}

func (p LintPosition) String() string {
	return fmt.Sprintf("%d:%d", p.Line+1, p.Ch+1)
}

type severity string

const (
	severityError   severity = "error"
	severityWraning severity = "warning"
)

type LintMessage struct {
	After    string       `json:"after"`
	Before   string       `json:"before"`
	From     LintPosition `json:"from"`
	To       LintPosition `json:"to"`
	Index    int          `json:"index"`
	IndexTo  int          `json:"index_to"`
	Message  string       `json:"message"`
	Severity severity     `json:"severity"`
}

type LintResult struct {
	Messages []LintMessage `json:"messages"`
	Status   lintStatus    `json:"status"`
	Updated  int64         `json:"updated"`
}

type lintStatus string

const (
	statusDone       lintStatus = "done"
	statusProcessing lintStatus = "processing"
	statusFailed     lintStatus = "failed"
)

var (
	notFoundError       = errors.New("server returned response code: 400")
	lintProcessingError = errors.New("lint is processing")
	lintFailedError     = errors.New("lint failed")
)

func (c *client) LintResult(ctx context.Context, id string) (*LintResult, error) {
	for trial := 10; trial > 0; trial-- {
		r, err := c.lintResult(ctx, id)
		if err == nil {
			sort.Slice(r.Messages, func(i, j int) bool {
				m1, m2 := r.Messages[i], r.Messages[j]
				return m1.Index < m2.Index || m1.Index == m2.Index && m1.IndexTo < m2.IndexTo
			})
			return r, nil
		}
		if errors.Is(err, notFoundError) || errors.Is(err, lintProcessingError) {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("failed to lint for id: %s", id)
}

func (c *client) lintResult(ctx context.Context, id string) (*LintResult, error) {
	resp, err := c.c.Do(c.newRequest(ctx, http.MethodGet, fmt.Sprintf("lint/%s/", id), nil))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, notFoundError
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned response code: %d", resp.StatusCode)
	}

	var r LintResult
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.Status == statusProcessing {
		return nil, lintProcessingError
	}
	if r.Status == statusFailed {
		return nil, lintFailedError
	}
	return &r, nil
}
