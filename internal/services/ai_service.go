package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AIService handles AI-related operations
type AIService struct {
	openAIKey string
	client    *http.Client
}

// NewAIService creates a new AI service instance
func NewAIService() *AIService {
	return &AIService{
		openAIKey: os.Getenv("OPENAI_API_KEY"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SongSection represents a song section for AI generation
type SongSection struct {
	Title string `json:"title"`
	Key   string `json:"key"`
	Body  string `json:"body"`
}

// AIGenerationRequest represents the request for AI generation
type AIGenerationRequest struct {
	SongTitle string `json:"song_title"`
	Artist    string `json:"artist"`
	Prompt    string `json:"prompt"`
}

// AIGenerationResponse represents the response from AI generation
type AIGenerationResponse struct {
	SongTitle string        `json:"song_title"`
	Artist    string        `json:"artist"`
	Sections  []SongSection `json:"sections"`
}

// GenerateSongSections generates song sections using AI
func (s *AIService) GenerateSongSections(req *AIGenerationRequest) (*AIGenerationResponse, error) {
	// Add a 1-second delay to simulate processing time
	time.Sleep(1 * time.Second)

	// If no OpenAI key is configured, return sample data
	if s.openAIKey == "" {
		return s.generateSampleSections(req.SongTitle, req.Artist), nil
	}

	// Create the prompt for ChatGPT
	prompt := fmt.Sprintf(`You are a music expert and songwriter. Generate song sections for "%s" by %s. 

Please provide the response in the following JSON format:

{
  "song_title": "%s",
  "artist": "%s",
  "sections": [
    {
      "title": "Intro",
      "key": "C",
      "body": "**Intro (4 bars)**\\n\\nC - Am - F - G\\n\\n*Simple chord progression to establish the key*"
    },
    {
      "title": "Verse 1",
      "key": "C",
      "body": "**Verse 1**\\n\\nC           Am          F           G\\nThis is the first verse of our song\\n\\nC           Am          F           G\\nWith chords written above the lyrics\\n\\n*Play with a gentle strumming pattern*"
    }
  ]
}

Please generate realistic song sections with proper chord progressions, lyrics, and musical instructions. Use Markdown formatting in the body content. Include 4-6 sections like Intro, Verse 1, Chorus, Verse 2, Bridge, and Outro.`, req.SongTitle, req.Artist, req.SongTitle, req.Artist)

	// Call OpenAI API
	openAIReq := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a music expert and songwriter. Generate song sections in the exact JSON format requested.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonData, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	// Make request to OpenAI
	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.openAIKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make OpenAI request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	// Parse OpenAI response
	var openAIResp map[string]interface{}
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Extract the content from the response
	choices, ok := openAIResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid content format")
	}

	// Parse the JSON content from the AI response
	var aiResponse AIGenerationResponse
	if err := json.Unmarshal([]byte(content), &aiResponse); err != nil {
		// If parsing fails, return sample data
		return s.generateSampleSections(req.SongTitle, req.Artist), nil
	}

	return &aiResponse, nil
}

// generateSampleSections creates sample song sections when AI is not available
func (s *AIService) generateSampleSections(songTitle, artist string) *AIGenerationResponse {
	return &AIGenerationResponse{
		SongTitle: songTitle,
		Artist:    artist,
		Sections: []SongSection{
			{
				Title: "Intro",
				Key:   "C",
				Body:  "**Intro (4 bars)**\n\nC - Am - F - G\n\n*Simple chord progression to establish the key*",
			},
			{
				Title: "Verse 1",
				Key:   "C",
				Body:  fmt.Sprintf("**Verse 1**\n\nC           Am          F           G\nThis is the first verse of %s\n\nC           Am          F           G\nBy %s with chords above lyrics\n\n*Play with a gentle strumming pattern*", songTitle, artist),
			},
			{
				Title: "Chorus",
				Key:   "C",
				Body:  "**Chorus**\n\nF           C           G           Am\nThis is the chorus, it's the hook\n\nF           C           G           Am\nThat everyone will remember\n\n*Build up the energy here*",
			},
			{
				Title: "Verse 2",
				Key:   "C",
				Body:  "**Verse 2**\n\nC           Am          F           G\nSecond verse with different lyrics\n\nC           Am          F           G\nBut same chord progression as verse 1",
			},
			{
				Title: "Bridge",
				Key:   "Am",
				Body:  "**Bridge**\n\nAm          F           C           G\nBridge section changes the mood\n\nAm          F           C           G\nDifferent chord progression here\n\n*Play with more intensity*",
			},
			{
				Title: "Outro",
				Key:   "C",
				Body:  "**Outro**\n\nC - Am - F - G (repeat 2x)\n\n*Fade out gradually*\n\n**End on C**",
			},
		},
	}
}
