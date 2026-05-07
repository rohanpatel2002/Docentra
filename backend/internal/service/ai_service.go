package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

type AIService interface {
	GetEmbedding(ctx context.Context, text string) ([]float32, error)
	GetEmbeddings(ctx context.Context, text string) ([]ChunkResult, error)
}

type ChunkResult struct {
	Content   string    `json:"content"`
	Embedding []float32 `json:"embedding"`
}

type aiService struct {
	pythonPath string
	scriptPath string
}

func NewAIService(pythonPath, scriptPath string) AIService {
	return &aiService{
		pythonPath: pythonPath,
		scriptPath: scriptPath,
	}
}

type cliResponse struct {
	Chunks []ChunkResult `json:"chunks"`
}

func (s *aiService) GetEmbeddings(ctx context.Context, text string) ([]ChunkResult, error) {
	// Execute Python CLI command securely
	cmd := exec.CommandContext(ctx, s.pythonPath, s.scriptPath, "--text", text)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python execution failed: %w (output: %s)", err, string(output))
	}

	var resp cliResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse python output: %w", err)
	}

	return resp.Chunks, nil
}

func (s *aiService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	chunks, err := s.GetEmbeddings(ctx, text)
	if err != nil {
		return nil, err
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no embeddings returned from CLI")
	}
	return chunks[0].Embedding, nil
}
