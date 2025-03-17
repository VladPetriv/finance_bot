package gemini

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type gemini struct {
	client *genai.Client

	model string
}

// New creates a new instance of the gemini API.
func New(ctx context.Context, apiKey, model string) (*gemini, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("init gemini client: %w", err)
	}

	return &gemini{
		client: client,
		model:  model,
	}, nil
}

var errEmptyResponse = errors.New("empty response")

func (g *gemini) Execute(ctx context.Context, prompt string) (string, error) {
	response, err := g.client.GenerativeModel(g.model).GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("generate content through model: %w", err)
	}
	if len(response.Candidates) == 0 {
		return "", errEmptyResponse
	}

	var output string

candidateLoop:
	for _, candidate := range response.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				data, ok := part.(genai.Text)
				if !ok {
					continue
				}

				output += string(data)
			}

			break candidateLoop
		}
	}

	return output, nil
}

func (g *gemini) Close() error {
	if g.client != nil {
		return g.client.Close()
	}

	return nil
}
