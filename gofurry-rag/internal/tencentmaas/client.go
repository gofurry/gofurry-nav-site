package tencentmaas

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrNotConfigured = errors.New("tencent chat client is not configured")

type Client struct {
	baseURL         string
	apiKey          string
	model           string
	timeout         time.Duration
	temperature     float64
	topP            float64
	maxTokens       int
	reasoningEffort string
	client          *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionResult struct {
	Model            string
	Answer           string
	Reasoning        string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	ReasoningTokens  int
	CachedTokens     int
}

type requestPayload struct {
	Model           string    `json:"model"`
	Messages        []Message `json:"messages"`
	Stream          bool      `json:"stream"`
	Temperature     float64   `json:"temperature,omitempty"`
	TopP            float64   `json:"top_p,omitempty"`
	MaxTokens       int       `json:"max_tokens,omitempty"`
	ReasoningEffort string    `json:"reasoning_effort,omitempty"`
}

type completionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		TotalTokens         int `json:"total_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
		CompletionTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	} `json:"usage"`
}

type streamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		TotalTokens         int `json:"total_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
		CompletionTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	} `json:"usage"`
}

func New(baseURL, apiKey, model string, timeout time.Duration, temperature, topP float64, maxTokens int, reasoningEffort string) *Client {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		baseURL:         strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:          strings.TrimSpace(apiKey),
		model:           strings.TrimSpace(model),
		timeout:         timeout,
		temperature:     temperature,
		topP:            topP,
		maxTokens:       maxTokens,
		reasoningEffort: strings.TrimSpace(reasoningEffort),
		client:          &http.Client{Timeout: timeout},
	}
}

func (c *Client) Model() string {
	return c.model
}

func (c *Client) Configured() bool {
	return c != nil && c.baseURL != "" && c.apiKey != "" && c.model != ""
}

func (c *Client) Health(ctx context.Context) error {
	if !c.Configured() {
		return ErrNotConfigured
	}
	return nil
}

func (c *Client) Complete(ctx context.Context, messages []Message) (CompletionResult, error) {
	resp, err := c.doRequest(ctx, requestPayload{
		Model:           c.model,
		Messages:        messages,
		Stream:          false,
		Temperature:     c.temperature,
		TopP:            c.topP,
		MaxTokens:       c.maxTokens,
		ReasoningEffort: c.reasoningEffort,
	})
	if err != nil {
		return CompletionResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CompletionResult{}, readErrorResponse(resp.Body, resp.StatusCode)
	}

	var decoded completionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return CompletionResult{}, err
	}
	return completionFromResponse(decoded), nil
}

func (c *Client) Stream(ctx context.Context, messages []Message, onDelta func(string) error) (CompletionResult, error) {
	resp, err := c.doRequest(ctx, requestPayload{
		Model:           c.model,
		Messages:        messages,
		Stream:          true,
		Temperature:     c.temperature,
		TopP:            c.topP,
		MaxTokens:       c.maxTokens,
		ReasoningEffort: c.reasoningEffort,
	})
	if err != nil {
		return CompletionResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CompletionResult{}, readErrorResponse(resp.Body, resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	result := CompletionResult{}
	var answer strings.Builder
	var reasoning strings.Builder
	var dataLines []string

	flush := func() (bool, error) {
		if len(dataLines) == 0 {
			return false, nil
		}
		payload := strings.TrimSpace(strings.Join(dataLines, "\n"))
		dataLines = dataLines[:0]
		if payload == "" {
			return false, nil
		}
		if payload == "[DONE]" {
			return true, nil
		}
		var chunk streamResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return false, err
		}
		if chunk.Model != "" {
			result.Model = chunk.Model
		}
		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]
			if choice.Delta.Content != "" {
				answer.WriteString(choice.Delta.Content)
				if onDelta != nil {
					if err := onDelta(choice.Delta.Content); err != nil {
						return false, err
					}
				}
			}
			if choice.Delta.ReasoningContent != "" {
				reasoning.WriteString(choice.Delta.ReasoningContent)
			}
			if choice.FinishReason != "" && result.Model == "" {
				result.Model = c.model
			}
		}
		result.PromptTokens = chunk.Usage.PromptTokens
		result.CompletionTokens = chunk.Usage.CompletionTokens
		result.TotalTokens = chunk.Usage.TotalTokens
		result.CachedTokens = chunk.Usage.PromptTokensDetails.CachedTokens
		result.ReasoningTokens = chunk.Usage.CompletionTokensDetails.ReasoningTokens
		return false, nil
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return CompletionResult{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		switch {
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		case line == "":
			done, flushErr := flush()
			if flushErr != nil {
				return CompletionResult{}, flushErr
			}
			if done {
				result.Answer = answer.String()
				result.Reasoning = reasoning.String()
				if result.Model == "" {
					result.Model = c.model
				}
				return result, nil
			}
		}
		if errors.Is(err, io.EOF) {
			done, flushErr := flush()
			if flushErr != nil {
				return CompletionResult{}, flushErr
			}
			if !done {
				result.Answer = answer.String()
				result.Reasoning = reasoning.String()
				if result.Model == "" {
					result.Model = c.model
				}
			}
			return result, nil
		}
	}
}

func (c *Client) doRequest(ctx context.Context, payload requestPayload) (*http.Response, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if payload.Stream {
		req.Header.Set("Accept", "text/event-stream")
	}
	return c.client.Do(req)
}

func completionFromResponse(resp completionResponse) CompletionResult {
	result := CompletionResult{Model: resp.Model}
	if len(resp.Choices) > 0 {
		result.Answer = resp.Choices[0].Message.Content
		result.Reasoning = resp.Choices[0].Message.ReasoningContent
		if result.Answer == "" {
			result.Answer = result.Reasoning
		}
	}
	result.PromptTokens = resp.Usage.PromptTokens
	result.CompletionTokens = resp.Usage.CompletionTokens
	result.TotalTokens = resp.Usage.TotalTokens
	result.CachedTokens = resp.Usage.PromptTokensDetails.CachedTokens
	result.ReasoningTokens = resp.Usage.CompletionTokensDetails.ReasoningTokens
	return result
}

func readErrorResponse(body io.Reader, statusCode int) error {
	data, _ := io.ReadAll(body)
	message := strings.TrimSpace(string(data))
	if message == "" {
		return fmt.Errorf("tencent chat returned %d", statusCode)
	}
	var decoded struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &decoded); err == nil {
		switch {
		case decoded.Error.Message != "":
			message = decoded.Error.Message
		case decoded.Message != "":
			message = decoded.Message
		}
	}
	return fmt.Errorf("tencent chat returned %d: %s", statusCode, message)
}
