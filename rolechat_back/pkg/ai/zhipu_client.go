package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ZhipuClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	chatModel  string
	ttsModel   string
}

func NewZhipuClient(apiKey string) *ZhipuClient {
	return &ZhipuClient{
		httpClient: &http.Client{},
		apiKey:     apiKey,
		baseURL:    "https://open.bigmodel.cn/api/paas/v4",
		chatModel:  "glm-4",
		ttsModel:   "audio-01",
	}
}

type ChatMessage struct {
	Role    string
	Content string
}

type mmContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func (m ChatMessage) MarshalJSON() ([]byte, error) {
	aux := struct {
		Role    string      `json:"role"`
		Content []mmContent `json:"content"`
	}{Role: m.Role, Content: []mmContent{{Type: "text", Text: m.Content}}}
	return json.Marshal(aux)
}

type chatReq struct {
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages"`
	Stream           bool          `json:"stream,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	Temperature      float32       `json:"temperature,omitempty"`
	TopP             float32       `json:"top_p,omitempty"`
	DoSample         *bool         `json:"do_sample,omitempty"`
	WatermarkEnabled *bool         `json:"watermark_enabled,omitempty"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
			Audio   *struct {
				Data string `json:"data"`
			} `json:"audio,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *ZhipuClient) Chat(messages []ChatMessage) (string, error) {
	body, _ := json.Marshal(chatReq{Model: c.chatModel, Messages: messages, Temperature: 0.7})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("zhipu chat error: %s", string(data))
	}
	var cr chatResp
	if err := json.Unmarshal(data, &cr); err != nil {
		return "", err
	}
	if cr.Error != nil {
		return "", fmt.Errorf("zhipu api error: %s %s", cr.Error.Code, cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return cr.Choices[0].Message.Content, nil
}

func (c *ZhipuClient) ChatVoice(messages []ChatMessage) (string, string, error) {
	model := "glm-4-voice"
	doSample := true
	body, _ := json.Marshal(chatReq{Model: model, Messages: messages, Temperature: 0.8, TopP: 0.6, MaxTokens: 1024, DoSample: &doSample})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("zhipu chat voice error: %s", string(data))
	}
	var cr chatResp
	if err := json.Unmarshal(data, &cr); err != nil {
		return "", "", err
	}
	if cr.Error != nil {
		return "", "", fmt.Errorf("zhipu api error: %s %s", cr.Error.Code, cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", "", fmt.Errorf("empty choices")
	}
	msg := cr.Choices[0].Message
	audio := ""
	if msg.Audio != nil {
		audio = msg.Audio.Data
	}
	return msg.Content, audio, nil
}

type chatStreamResp struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (c *ZhipuClient) ChatStream(messages []ChatMessage) (<-chan string, <-chan error) {
	tokens := make(chan string, 8)
	errCh := make(chan error, 1)
	go func() {
		defer close(tokens)
		defer close(errCh)
		body, _ := json.Marshal(chatReq{Model: c.chatModel, Messages: messages, Temperature: 0.7, Stream: true})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		resp, err := c.httpClient.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			data, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("chat stream http %d: %s", resp.StatusCode, string(data))
			return
		}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" {
				break
			}
			var chunk chatStreamResp
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				tokens <- payload
				continue
			}
			for _, ch := range chunk.Choices {
				if ch.Delta.Content != "" {
					tokens <- ch.Delta.Content
				}
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			errCh <- err
		}
	}()
	return tokens, errCh
}

type ttsReq struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Voice  string `json:"voice,omitempty"`
	Format string `json:"format,omitempty"`
}

type ttsResp struct {
	AudioBase64 string `json:"audio_base64"`
	Error       *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *ZhipuClient) TextToSpeech(text, voice string) (string, error) {
	body, _ := json.Marshal(ttsReq{Model: c.ttsModel, Input: text, Voice: voice, Format: "mp3"})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/audio/tts", c.baseURL), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("tts error: %s", string(data))
	}
	var tr ttsResp
	if err := json.Unmarshal(data, &tr); err != nil {
		return "", err
	}
	if tr.Error != nil {
		return "", fmt.Errorf("tts api error: %s %s", tr.Error.Code, tr.Error.Message)
	}
	return tr.AudioBase64, nil
}
