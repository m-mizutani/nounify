package model

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/domain/types"
)

type MessageQueryInput struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Header map[string]string `json:"header"`
	Body   any               `json:"body"`
}

func NewMessageQueryInput(r *http.Request) (*MessageQueryInput, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, goerr.Wrap(types.ErrInvalidInput.Wrap(err)).With("method", r.Method).With("path", r.URL.Path)
	}

	var data any
	switch r.Header.Get("Content-Type") {
	case "application/json":
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, goerr.Wrap(types.ErrInvalidInput.Wrap(err)).
				With("method", r.Method).
				With("path", r.URL.Path).
				With("body", string(body))
		}

	default:
		data = string(body)
	}

	input := &MessageQueryInput{
		Method: r.Method,
		Path:   r.URL.Path,
		Header: map[string]string{},
		Body:   data,
	}

	for key := range r.Header {
		input.Header[key] = r.Header.Get(key)
	}

	return input, nil
}

type MessageQueryOutput struct {
	Messages []Message `json:"msg"`
}

type Message struct {
	Channel string         `json:"channel"`
	Color   string         `json:"color"`
	Title   string         `json:"title"`
	Body    string         `json:"body"`
	Fields  []MessageField `json:"fields"`
	Icon    string         `json:"icon"`
	Emoji   string         `json:"emoji"`
}

type MessageField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Link  string `json:"link"`
}

type AuthQueryInput struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Header map[string]string `json:"header"`
	Auth   struct {
		GitHub GitHubAuth     `json:"github"`
		Google map[string]any `json:"google"`
	} `json:"auth"`
}

type AuthQueryOutput struct {
	Allow bool `json:"allow"`
}
