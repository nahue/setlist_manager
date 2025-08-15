# AI Song Sections Feature

This feature allows users to automatically generate song sections using AI. The system can integrate with OpenAI's ChatGPT API or fall back to sample data when no API key is configured.

## Features

- **AI-Powered Generation**: Generate song sections using ChatGPT
- **Fallback Mode**: Works without API key using sample data
- **Markdown Support**: Generated content supports Markdown formatting
- **Musician-Friendly**: Includes chord progressions, playing instructions, and musical context
- **Real-time Updates**: Sections are created and displayed immediately

## Setup

### Option 1: With OpenAI API (Recommended)

1. Get an OpenAI API key from [OpenAI Platform](https://platform.openai.com/)
2. Set the environment variable:
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```
3. Or add it to your `.env` file:
   ```
   OPENAI_API_KEY=your-api-key-here
   ```

### Option 2: Without API Key

The feature will work with sample data when no API key is configured. This is perfect for testing or development.

## How to Use

1. **Navigate to a Song**: Go to any song's details page
2. **Find the Song Sections**: Scroll down to the "Song Sections" area
3. **Click "Generate with AI"**: Look for the purple "Generate with AI" button
4. **Wait for Generation**: The system will create song sections automatically
5. **Review and Edit**: The generated sections will appear and can be edited

## Generated Content

The AI generates the following types of sections:

- **Intro**: Opening chord progression and musical setup
- **Verse 1/2**: Main song verses with lyrics and chords
- **Chorus**: The hook section with full energy
- **Bridge**: Musical transition section
- **Outro**: Closing section with fade-out instructions

Each section includes:
- **Title**: Standard section name (Intro, Verse, Chorus, etc.)
- **Key**: Musical key (C, Am, F#, etc.)
- **Body**: Content with:
  - Chord progressions
  - Lyrics (when applicable)
  - Playing instructions
  - Markdown formatting

## Technical Details

### API Endpoint

```
POST /api/songs/{songID}/sections/generate-ai
```

### Request Body

```json
{
  "prompt": "Generated prompt for AI",
  "song_title": "Song Name",
  "artist": "Artist Name"
}
```

### Response

Returns HTML with the updated song sections component.

### AI Service Integration

The system uses the `AIService` which:
- Checks for `OPENAI_API_KEY` environment variable
- Makes requests to OpenAI's ChatGPT API
- Falls back to sample data if no key is configured
- Handles API errors gracefully

## Customization

### Modifying the AI Prompt

Edit the prompt in `internal/services/ai_service.go`:

```go
prompt := fmt.Sprintf(`You are a music expert and songwriter. Generate song sections for "%s" by %s. 
// ... rest of prompt
```

### Adding New Section Types

Update the prompt to include new section types or modify the sample data generation.

### Changing AI Model

Modify the model in the OpenAI request:

```go
"model": "gpt-4", // or "gpt-3.5-turbo"
```

## Error Handling

The system handles various error scenarios:

- **No API Key**: Falls back to sample data
- **API Errors**: Shows error notification and continues with sample data
- **Network Issues**: Graceful degradation with user feedback
- **Invalid Responses**: Parsing errors fall back to sample data

## Security

- API keys are stored as environment variables
- No API keys are logged or stored in the database
- All requests are authenticated through the existing auth system
- Band membership is verified before allowing AI generation

## Performance

- AI requests have a 30-second timeout
- Generated sections are cached in the database
- UI updates are immediate after generation
- Loading states provide user feedback

## Troubleshooting

### AI Generation Not Working

1. Check if `OPENAI_API_KEY` is set correctly
2. Verify the API key has sufficient credits
3. Check network connectivity
4. Review server logs for error messages

### Sample Data Instead of AI

This is expected behavior when:
- No API key is configured
- API key is invalid
- API service is unavailable
- Rate limits are exceeded

### Custom Prompts

To use custom prompts, modify the JavaScript in `templates/song_sections.templ`:

```javascript
const prompt = `Your custom prompt here...`;
```

## Future Enhancements

Potential improvements:
- Support for multiple AI providers
- Custom prompt templates per band
- Genre-specific generation
- Chord progression analysis
- Integration with music theory databases

