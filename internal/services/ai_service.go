package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

// SongInfo represents song metadata for the cheatsheet
type SongInfo struct {
	Title         string `json:"title"`
	Artist        string `json:"artist"`
	OriginalKey   string `json:"original_key"`
	Tempo         string `json:"tempo"`
	TimeSignature string `json:"time_signature"`
	Duration      string `json:"duration"`
}

// SongSection represents a song section for AI generation
type SongSection struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	Body string `json:"body"`
}

// AIGenerationRequest represents the request for AI generation
type AIGenerationRequest struct {
	SongTitle string `json:"song_title"`
	Artist    string `json:"artist"`
	Prompt    string `json:"prompt"`
	Key       string `json:"key"`
}

// AIGenerationResponse represents the response from AI generation
type AIGenerationResponse struct {
	SongInfo SongInfo      `json:"song_info"`
	Sections []SongSection `json:"sections"`
}

// GenerateSongSections generates song sections using AI
func (s *AIService) GenerateSongSections(req *AIGenerationRequest) (*AIGenerationResponse, error) {
	// Add a 1-second delay to simulate processing time
	time.Sleep(1 * time.Second)

	// If no OpenAI key is configured, return sample data
	if s.openAIKey == "" {
		return s.generateSampleSections(req.SongTitle, req.Artist), nil
	}

	// Create the prompt for ChatGPT using a flexible format
	prompt := fmt.Sprintf(`Generate song sections for "%s" by %s in the key of %s. 

Return JSON in this exact format:

{"song_info":{"title":"%s","artist":"%s","original_key":"%s","tempo":"[BPM and feel]","time_signature":"[4/4, 3/4, etc.]","duration":"[mm:ss]"},"sections":[{"name":"[section_name]","key":"%s","body":"**Lyrics:** [Complete lyrics]\\n\\n**Notes:** [Performance notes]"}]}

IMPORTANT: For each section, provide the COMPLETE lyrics. Do not use placeholders like [Instrumental], [...], or partial lyrics. Write out the full lyrics for each section. For instrumental sections, write "(Instrumental)" as the lyrics.`, req.SongTitle, req.Artist, req.Key, req.SongTitle, req.Artist, req.Key, req.Key)

	// Call OpenAI API
	openAIReq := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a music expert and band practice coach. Generate comprehensive band practice cheatsheets in the exact JSON format requested, focusing on performance aspects rather than technical music theory.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  4096,
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

	// Clean the content by removing markdown code blocks if present
	cleanedContent := content
	if strings.HasPrefix(strings.TrimSpace(content), "```") {
		lines := strings.Split(content, "\n")
		if len(lines) > 2 {
			// Remove first line (```json) and last line (```)
			cleanedContent = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	// If the JSON is truncated, try to complete it
	if !strings.HasSuffix(strings.TrimSpace(cleanedContent), "}") {
		// Find the last complete object and close it
		lastBrace := strings.LastIndex(cleanedContent, "}")
		if lastBrace > 0 {
			cleanedContent = cleanedContent[:lastBrace+1]
		}
	}

	// Parse the JSON content from the AI response
	var aiResponse AIGenerationResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		// If parsing fails, return the error with the content for debugging
		return nil, fmt.Errorf("failed to parse AI response JSON: %w\nAI Response Content: %s", err, content)
	}

	// Convert escaped newlines to actual newlines in the body content
	for i := range aiResponse.Sections {
		aiResponse.Sections[i].Body = strings.ReplaceAll(aiResponse.Sections[i].Body, "\\n", "\n")
	}

	fmt.Println("AI Prompt:", prompt)
	fmt.Println("AI Response:", aiResponse)
	return &aiResponse, nil
}

// generateSampleSections creates sample song sections when AI is not available
func (s *AIService) generateSampleSections(songTitle, artist string) *AIGenerationResponse {
	return &AIGenerationResponse{
		SongInfo: SongInfo{
			Title:         songTitle,
			Artist:        artist,
			OriginalKey:   "C",
			Tempo:         "120 BPM, driving rock",
			TimeSignature: "4/4",
			Duration:      "03:00",
		},
		Sections: []SongSection{
			{
				Name: "intro",
				Key:  "C",
				Body: "**Lyrics:** [Instrumental intro]\n\n**Notes:** Gentle arpeggiated chords, establish the mood, everyone enters together",
			},
			{
				Name: "verse_1",
				Key:  "C",
				Body: "**Lyrics:** This is the first verse of our song\nWith chords written above the lyrics\n\n**Notes:** Clean strumming, medium volume, building energy, clear vocal delivery",
			},
			{
				Name: "chorus",
				Key:  "C",
				Body: "**Lyrics:** This is the chorus, it's the hook\nThat everyone will remember\n\n**Notes:** High energy, power chords, full band, anthemic feel",
			},
			{
				Name: "verse_2",
				Key:  "C",
				Body: "**Lyrics:** Second verse with different lyrics\nBut same chord progression as verse 1\n\n**Notes:** More intensity than verse 1, add guitar fills, stronger vocal delivery",
			},
			{
				Name: "bridge",
				Key:  "Am",
				Body: "**Lyrics:** Bridge section changes the mood\nDifferent chord progression here\n\n**Notes:** Different key, emotional intensity, fingerpicking, dramatic pause",
			},
			{
				Name: "outro",
				Key:  "C",
				Body: "**Lyrics:** Final lyrics for the outro\nEnding with a gentle fade\n\n**Notes:** Gradual fade out, sustained chords, soft ending",
			},
		},
	}
}
