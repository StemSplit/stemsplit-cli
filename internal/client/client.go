package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const baseURL = "https://stemsplit.io/api/v1"

// Client wraps the StemSplit HTTP API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// --- Request / Response types ---

type UploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType,omitempty"`
}

type UploadResponse struct {
	UploadURL string `json:"uploadUrl"`
	UploadKey string `json:"uploadKey"`
	ExpiresAt string `json:"expiresAt"`
}

type CreateJobRequest struct {
	UploadKey    string                 `json:"uploadKey,omitempty"`
	SourceURL    string                 `json:"sourceUrl,omitempty"`
	FileName     string                 `json:"fileName,omitempty"`
	OutputType   string                 `json:"outputType,omitempty"`
	Quality      string                 `json:"quality,omitempty"`
	OutputFormat string                 `json:"outputFormat,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type JobOutput struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expiresAt"`
}

type JobInput struct {
	FileName        string `json:"fileName"`
	DurationSeconds int    `json:"durationSeconds"`
	FileSizeBytes   int    `json:"fileSizeBytes"`
}

type JobOptions struct {
	OutputType   string `json:"outputType"`
	Quality      string `json:"quality"`
	OutputFormat string `json:"outputFormat"`
}

type Job struct {
	ID               string                `json:"id"`
	Status           string                `json:"status"`
	Progress         int                   `json:"progress"`
	CreatedAt        string                `json:"createdAt"`
	StartedAt        *string               `json:"startedAt"`
	CompletedAt      *string               `json:"completedAt"`
	Input            JobInput              `json:"input"`
	Options          JobOptions            `json:"options"`
	Outputs          map[string]*JobOutput `json:"outputs"`
	CreditsCharged   int                   `json:"creditsCharged"`
	CreditsRequired  int                   `json:"creditsRequired"`
	EstimatedSeconds int                   `json:"estimatedSeconds"`
	ErrorMessage     *string               `json:"errorMessage"`
	ExpiresAt        *string               `json:"expiresAt"`
}

type JobsListResponse struct {
	Jobs       []Job `json:"jobs"`
	Pagination struct {
		Total   int  `json:"total"`
		Limit   int  `json:"limit"`
		Offset  int  `json:"offset"`
		HasMore bool `json:"hasMore"`
	} `json:"pagination"`
}

type BalanceResponse struct {
	BalanceSeconds   int    `json:"balanceSeconds"`
	BalanceMinutes   int    `json:"balanceMinutes"`
	BalanceFormatted string `json:"balanceFormatted"`
	UpdatedAt        string `json:"updatedAt"`
}

// --- API methods ---

func (c *Client) GetUploadURL(filename, contentType string) (*UploadResponse, error) {
	body := UploadRequest{Filename: filename, ContentType: contentType}
	var resp UploadResponse
	if err := c.do("POST", "/upload", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UploadFile sends the file bytes to the presigned URL via PUT (no auth header).
func (c *Client) UploadFile(uploadURL, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", uploadURL, f)
	if err != nil {
		return err
	}
	req.ContentLength = stat.Size()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) CreateJob(req *CreateJobRequest) (*Job, error) {
	var job Job
	if err := c.do("POST", "/jobs", req, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *Client) GetJob(id string) (*Job, error) {
	var job Job
	if err := c.do("GET", "/jobs/"+id, nil, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *Client) ListJobs(limit, offset int) (*JobsListResponse, error) {
	endpoint := fmt.Sprintf("/jobs?limit=%s&offset=%s",
		url.QueryEscape(strconv.Itoa(limit)),
		url.QueryEscape(strconv.Itoa(offset)),
	)
	var resp JobsListResponse
	if err := c.do("GET", endpoint, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetBalance() (*BalanceResponse, error) {
	var resp BalanceResponse
	if err := c.do("GET", "/balance", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DownloadFile downloads the URL to destPath.
func (c *Client) DownloadFile(downloadURL, destPath string) error {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("download error %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// --- Internal HTTP helper ---

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) do(method, path string, body interface{}, out interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "stemsplit-cli/"+version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var apiErr apiError
		if jsonErr := json.Unmarshal(respBytes, &apiErr); jsonErr == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("%s", apiErr.Error.Message)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	if out != nil {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}
	return nil
}

// version is set at build time via ldflags.
var version = "dev"
