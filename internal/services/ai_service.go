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

// SongContentRequest represents the request for song content generation
type SongContentRequest struct {
	SongTitle string `json:"song_title"`
	Artist    string `json:"artist"`
	Key       string `json:"key"`
	Tempo     *int   `json:"tempo"`
}

// SongContentResponse represents the response from song content generation
type SongContentResponse struct {
	Content string `json:"content"`
}

// GenerateSongContent generates song content using AI for band practice
func (s *AIService) GenerateSongContent(req *SongContentRequest) (*SongContentResponse, error) {
	// Add a 1-second delay to simulate processing time
	time.Sleep(1 * time.Second)

	// If no OpenAI key is configured, return sample data
	if s.openAIKey == "" {
		return s.generateSampleContent(req.SongTitle, req.Artist, req.Key, req.Tempo), nil
	}

	// Create the prompt for ChatGPT
	tempoStr := "medium tempo"
	if req.Tempo != nil {
		tempoStr = fmt.Sprintf("%d BPM", *req.Tempo)
	}

	prompt := fmt.Sprintf(`Generate a comprehensive band practice cheatsheet for "%s" by %s in the key of %s at %s.

The content should be formatted in Markdown and include:

1. **Song Structure** - Clear section breakdown (Intro, Verse, Chorus, Bridge, etc.)
2. **Complete Lyrics** - Full lyrics for each section (no placeholders like [...])
3. **Chord Progressions** - Chords written above lyrics where appropriate
4. **Performance Notes** - Specific hints for band members including:
   - Dynamics (when to play soft/loud)
   - Rhythmic patterns
   - Guitar techniques (strumming, fingerpicking, etc.)
   - Bass lines and drum patterns
   - Vocal delivery tips
   - How sections connect and flow
5. **Musical Feel** - Overall mood and energy of each section

Format the response as clean Markdown with clear headers, bullet points, and organized sections. Focus on practical information that helps band members perform the song effectively.

IMPORTANT: Include the COMPLETE lyrics for each section. Do not use placeholders or partial lyrics.`, req.SongTitle, req.Artist, req.Key, tempoStr)

	// Call OpenAI API
	openAIReq := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a music expert and band practice coach. Generate comprehensive band practice cheatsheets in Markdown format, focusing on practical performance aspects rather than technical music theory. Always include complete lyrics and specific performance hints for each band member.",
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

	return &SongContentResponse{Content: content}, nil
}

// generateSampleContent creates sample song content when AI is not available
func (s *AIService) generateSampleContent(songTitle, artist, key string, tempo *int) *SongContentResponse {
	tempoStr := "medium tempo"
	if tempo != nil {
		tempoStr = fmt.Sprintf("%d BPM", *tempo)
	}

	sampleContent := fmt.Sprintf(`# %s - %s

**Key:** %s | **Tempo:** %s

## Song Structure
- Intro
- Verse 1
- Chorus
- Verse 2
- Chorus
- Bridge
- Final Chorus
- Outro

## Intro
**Chords:** %s - Am - F - G

**Notes:** Gentle arpeggiated chords, establish the mood, everyone enters together. Start soft and build energy.

## Verse 1
**Chords:** %s - Am - F - G

**Lyrics:**
This is the first verse of our song
With chords written above the lyrics
Building the story and the mood
Setting up for the chorus to come

**Performance Notes:**
- Clean strumming, medium volume
- Building energy throughout the verse
- Clear vocal delivery
- Bass follows chord roots
- Drums: steady 4/4 pattern

## Chorus
**Chords:** F - G - Am - %s

**Lyrics:**
This is the chorus, it's the hook
That everyone will remember
The part that gets stuck in your head
The moment when we all sing together

**Performance Notes:**
- High energy, power chords
- Full band, anthemic feel
- Strong vocal harmonies
- Guitar: palm-muted power chords
- Bass: driving eighth notes
- Drums: full kit, strong backbeat

## Verse 2
**Chords:** %s - Am - F - G

**Lyrics:**
Second verse with different lyrics
But same chord progression as verse 1
Building on the story we started
Taking it to a deeper level

**Performance Notes:**
- More intensity than verse 1
- Add guitar fills between lines
- Stronger vocal delivery
- Bass: more melodic lines
- Drums: add hi-hat variations

## Bridge
**Chords:** Am - F - C - G

**Lyrics:**
Bridge section changes the mood
Different chord progression here
Taking the song in a new direction
Before we return to the chorus

**Performance Notes:**
- Different key, emotional intensity
- Fingerpicking guitar pattern
- Dramatic pause before final chorus
- Bass: walking bass line
- Drums: half-time feel

## Final Chorus
**Chords:** F - G - Am - %s

**Lyrics:**
This is the chorus, it's the hook
That everyone will remember
The part that gets stuck in your head
The moment when we all sing together

**Performance Notes:**
- Maximum energy and volume
- Full band, everyone singing
- Guitar: full power chords
- Bass: driving and melodic
- Drums: full kit, strong fills

## Outro
**Chords:** %s - Am - F - G

**Lyrics:**
Final lyrics for the outro
Ending with a gentle fade
Taking the energy down slowly
Until we're just a whisper

**Performance Notes:**
- Gradual fade out
- Sustained chords
- Soft ending
- Guitar: gentle arpeggios
- Bass: long notes
- Drums: soft brushes`, songTitle, artist, key, tempoStr, key, key, key, key, key, key)

	return &SongContentResponse{Content: sampleContent}
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
