package obsidian

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Config holds the Obsidian Local REST API connection settings.
type Config struct {
	APIKey  string
	BaseURL string // e.g. "http://127.0.0.1:27124"
}

// ConfigFromEnv reads Config from environment variables.
// OBSIDIAN_BASE_URL sets the base URL (default: http://127.0.0.1:27124).
func ConfigFromEnv() Config {
	base := os.Getenv("OBSIDIAN_BASE_URL")
	if base == "" {
		base = "http://127.0.0.1:27124"
	}
	base = strings.TrimRight(base, "/")
	if _, err := url.Parse(base); err != nil {
		log.Fatalf("invalid OBSIDIAN_BASE_URL %q: %v", base, err)
	}
	return Config{
		APIKey:  os.Getenv("OBSIDIAN_API_KEY"),
		BaseURL: base,
	}
}

// Client is an HTTP client for the Obsidian Local REST API.
type Client struct {
	cfg        Config
	httpClient *http.Client
}

// New constructs a Client from cfg.
func New(cfg Config) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   9 * time.Second,
		},
	}
}

type apiErrBody struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

// do performs an HTTP request. Non-2xx responses become errors; Obsidian's
// errorCode and message are surfaced when available.
func (c *Client) do(method, path string, headers map[string]string, body []byte) ([]byte, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, c.cfg.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e apiErrBody
		if json.Unmarshal(data, &e) == nil && e.Message != "" {
			return nil, fmt.Errorf("Error %d: %s", e.ErrorCode, e.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// ListFilesInVault returns all file paths in the vault root.
func (c *Client) ListFilesInVault() ([]string, error) {
	data, err := c.do(http.MethodGet, "/vault/", nil, nil)
	if err != nil {
		return nil, err
	}
	var r struct {
		Files []string `json:"files"`
	}
	return r.Files, json.Unmarshal(data, &r)
}

// ListFilesInDir returns all file paths under dirpath.
func (c *Client) ListFilesInDir(dirpath string) ([]string, error) {
	data, err := c.do(http.MethodGet, "/vault/"+dirpath+"/", nil, nil)
	if err != nil {
		return nil, err
	}
	var r struct {
		Files []string `json:"files"`
	}
	return r.Files, json.Unmarshal(data, &r)
}

// GetFileContents returns the raw text content of a vault file.
func (c *Client) GetFileContents(filepath string) (string, error) {
	data, err := c.do(http.MethodGet, "/vault/"+filepath, nil, nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetBatchFileContents returns the contents of multiple files concatenated
// with markdown headers.
func (c *Client) GetBatchFileContents(filepaths []string) string {
	var sb strings.Builder
	for _, fp := range filepaths {
		content, err := c.GetFileContents(fp)
		if err != nil {
			fmt.Fprintf(&sb, "# %s\n\nError reading file: %s\n\n---\n\n", fp, err)
		} else {
			fmt.Fprintf(&sb, "# %s\n\n%s\n\n---\n\n", fp, content)
		}
	}
	return sb.String()
}

// Search performs a simple full-text search.
func (c *Client) Search(query string, contextLength int) (json.RawMessage, error) {
	path := fmt.Sprintf("/search/simple/?query=%s&contextLength=%d",
		url.QueryEscape(query), contextLength)
	data, err := c.do(http.MethodPost, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// AppendContent appends text to a vault file.
func (c *Client) AppendContent(filepath, content string) error {
	_, err := c.do(http.MethodPost, "/vault/"+filepath,
		map[string]string{"Content-Type": "text/markdown; charset=utf-8"},
		[]byte(content))
	return err
}

// PatchContent inserts content relative to a heading, block, or frontmatter
// field. If the server returns error 40080 for a bare heading name (no "::"),
// it auto-qualifies by parsing the file's heading hierarchy.
func (c *Client) PatchContent(filepath, operation, targetType, target, content string) error {
	err := c.patchRaw(filepath, operation, targetType, target, content)
	if err == nil {
		return nil
	}
	if targetType != "heading" || strings.Contains(target, "::") || !strings.Contains(err.Error(), "Error 40080") {
		return err
	}
	fileContent, fetchErr := c.GetFileContents(filepath)
	if fetchErr != nil {
		return err
	}
	candidates := FindHeadingPaths(fileContent, target)
	switch len(candidates) {
	case 1:
		return c.patchRaw(filepath, operation, targetType, candidates[0], content)
	case 0:
		return err
	default:
		return fmt.Errorf("ambiguous heading %q; candidates: %s — use '::' qualified path",
			target, strings.Join(candidates, ", "))
	}
}

func (c *Client) patchRaw(filepath, operation, targetType, target, content string) error {
	_, err := c.do(http.MethodPatch, "/vault/"+filepath,
		map[string]string{
			"Content-Type": "text/markdown",
			"Operation":    operation,
			"Target-Type":  targetType,
			"Target":       url.PathEscape(target),
		},
		[]byte(content))
	return err
}

// PutContent overwrites a vault file with content.
func (c *Client) PutContent(filepath, content string) error {
	_, err := c.do(http.MethodPut, "/vault/"+filepath,
		map[string]string{"Content-Type": "text/markdown; charset=utf-8"},
		[]byte(content))
	return err
}

// DeleteFile deletes a vault file.
func (c *Client) DeleteFile(filepath string) error {
	_, err := c.do(http.MethodDelete, "/vault/"+filepath, nil, nil)
	return err
}

// OpenInUI opens a file in the Obsidian UI. If the file does not exist,
// Obsidian creates a new document at the given path.
func (c *Client) OpenInUI(filepath string) error {
	_, err := c.do(http.MethodPost, "/open/"+filepath, nil, nil)
	return err
}

// SearchJSON executes a JsonLogic query against the vault.
func (c *Client) SearchJSON(query any) (json.RawMessage, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	data, err := c.do(http.MethodPost, "/search/",
		map[string]string{"Content-Type": "application/vnd.olrapi.jsonlogic+json"},
		body)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// SearchByTag returns paths of notes with the given tag, optionally scoped
// to a vault subdirectory.
func (c *Client) SearchByTag(tag, dirpath string) ([]string, error) {
	tagClause := map[string]any{"in": []any{tag, map[string]any{"var": "tags"}}}
	var query map[string]any
	if dirpath != "" {
		prefix := strings.TrimRight(dirpath, "/") + "/"
		query = map[string]any{
			"and": []any{
				tagClause,
				map[string]any{"glob": []any{prefix + "*", map[string]any{"var": "path"}}},
			},
		}
	} else {
		query = tagClause
	}
	data, err := c.SearchJSON(query)
	if err != nil {
		return nil, err
	}
	var results []struct {
		Filename string `json:"filename"`
	}
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	paths := make([]string, len(results))
	for i, r := range results {
		paths[i] = r.Filename
	}
	return paths, nil
}

// GetFrontmatter returns parsed YAML frontmatter as a JSON object.
// Returns {} for notes without frontmatter.
func (c *Client) GetFrontmatter(filepath string) (json.RawMessage, error) {
	data, err := c.do(http.MethodGet, "/vault/"+filepath,
		map[string]string{"Accept": "application/vnd.olrapi.note+json"},
		nil)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Frontmatter json.RawMessage `json:"frontmatter"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if len(payload.Frontmatter) == 0 || string(payload.Frontmatter) == "null" {
		return json.RawMessage("{}"), nil
	}
	return payload.Frontmatter, nil
}

// GetPeriodicNote returns the content or metadata of the current periodic note.
func (c *Client) GetPeriodicNote(period, noteType string) (string, error) {
	headers := map[string]string{}
	if noteType == "metadata" {
		headers["Accept"] = "application/vnd.olrapi.note+json"
	}
	data, err := c.do(http.MethodGet, "/periodic/"+period+"/", headers, nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetRecentPeriodicNotes returns recently created periodic notes.
func (c *Client) GetRecentPeriodicNotes(period string, limit int, includeContent bool) (json.RawMessage, error) {
	path := fmt.Sprintf("/periodic/%s/recent?limit=%d&includeContent=%v", period, limit, includeContent)
	data, err := c.do(http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetRecentChanges queries recently modified files via DQL.
func (c *Client) GetRecentChanges(limit, days int) (json.RawMessage, error) {
	dql := fmt.Sprintf(
		"TABLE file.mtime\nWHERE file.mtime >= date(today) - dur(%d days)\nSORT file.mtime DESC\nLIMIT %d",
		days, limit)
	data, err := c.do(http.MethodPost, "/search/",
		map[string]string{"Content-Type": "application/vnd.olrapi.dataview.dql+txt"},
		[]byte(dql))
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
