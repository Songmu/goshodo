package shodo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const userAgentBase = "Songmu-shodo/%s (+https://github.com/Songmu/shodo)"

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
		return nil, fmt.Errorf("no shodo api root")
	}
	token, ok := os.LookupEnv("SHODO_API_TOKEN")
	if !ok {
		return nil, fmt.Errorf("no shodo api token")
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
func (c *client) createLint(ctx context.Context, body string) (string, error) {
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

type pos struct {
	Ch   int `json:"ch"`
	Line int `json:"line"`
}

func (p pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line+1, p.Ch+1)
}

type severity string

const (
	severityError   severity = "error"
	severityWraning severity = "warning"
)

type message struct {
	After    string   `json:"after"`
	Before   string   `json:"before"`
	From     pos      `json:"from"`
	To       pos      `json:"to"`
	Index    int      `json:"index"`
	IndexTo  int      `json:"index_to"`
	Message  string   `json:"message"`
	Severity severity `json:"severity"`
}

type response struct {
	Messages []message `json:"messages"`
	Status   string    `json:"status"`
	Updated  int64     `json:"updated"`
}

func (c *client) lintResult(ctx context.Context, id string) (*response, error) {
	resp, err := c.c.Do(c.newRequest(ctx, http.MethodGet, fmt.Sprintf("lint/%s/", id), nil))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned response code: %d", resp.StatusCode)
	}

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}
