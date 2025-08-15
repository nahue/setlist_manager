# ChatGPT Prompt for Generating Song Sections

You are a music expert and songwriter. I need you to generate song sections for a song that will be used in a setlist management application. 

## Song Section Model Structure

Each song section must have the following fields:
- **Title**: The name of the section (e.g., "Verse 1", "Chorus", "Bridge", "Intro", "Outro", "Pre-Chorus", "Solo", etc.)
- **Key**: The musical key for this section (optional, can be empty)
- **Body**: The content of the section (lyrics, chords, notes, etc.) - supports Markdown formatting

## Key Requirements

1. **Title Format**: Use standard section names like:
   - Intro
   - Verse 1, Verse 2, Verse 3, etc.
   - Pre-Chorus
   - Chorus
   - Bridge
   - Solo
   - Outro
   - Instrumental
   - Breakdown

2. **Key Format**: Use standard musical notation:
   - Major keys: C, C#, D, D#, E, F, F#, G, G#, A, A#, B
   - Minor keys: Cm, C#m, Dm, D#m, Em, Fm, F#m, Gm, G#m, Am, A#m, Bm
   - Can be empty if no specific key is needed

3. **Body Content**: 
   - Can include lyrics, chord progressions, notes, or any musical content
   - Supports Markdown formatting for better presentation
   - Can include chord charts, tablature, or musical notation
   - Should be descriptive and helpful for musicians

4. **Position**: Sections should be in logical order (Intro → Verse → Chorus → etc.)

## Output Format

Please provide the song sections in the following JSON format:

```json
{
  "song_title": "Song Name",
  "artist": "Artist Name",
  "sections": [
    {
      "title": "Intro",
      "key": "C",
      "body": "**Intro (4 bars)**\n\nC - Am - F - G\n\n*Simple chord progression to establish the key*"
    },
    {
      "title": "Verse 1",
      "key": "C",
      "body": "**Verse 1**\n\nC           Am          F           G\nThis is the first verse of our song\n\nC           Am          F           G\nWith chords written above the lyrics\n\n*Play with a gentle strumming pattern*"
    },
    {
      "title": "Chorus",
      "key": "C",
      "body": "**Chorus**\n\nF           C           G           Am\nThis is the chorus, it's the hook\n\nF           C           G           Am\nThat everyone will remember\n\n*Build up the energy here*"
    },
    {
      "title": "Verse 2",
      "key": "C",
      "body": "**Verse 2**\n\nC           Am          F           G\nSecond verse with different lyrics\n\nC           Am          F           G\nBut same chord progression as verse 1"
    },
    {
      "title": "Bridge",
      "key": "Am",
      "body": "**Bridge**\n\nAm          F           C           G\nBridge section changes the mood\n\nAm          F           C           G\nDifferent chord progression here\n\n*Play with more intensity*"
    },
    {
      "title": "Outro",
      "key": "C",
      "body": "**Outro**\n\nC - Am - F - G (repeat 2x)\n\n*Fade out gradually*\n\n**End on C**"
    }
  ]
}
```

## Instructions

1. **Generate realistic song sections** that follow typical song structure
2. **Include appropriate keys** for each section (can vary between sections)
3. **Use Markdown formatting** in the body content for better readability
4. **Provide helpful musical context** in the body (chord progressions, playing instructions, etc.)
5. **Ensure logical flow** from one section to the next
6. **Make content musician-friendly** with clear chord charts and instructions

## Example Request

"Generate song sections for 'Wonderwall' by Oasis"

## Example Response

```json
{
  "song_title": "Wonderwall",
  "artist": "Oasis",
  "sections": [
    {
      "title": "Intro",
      "key": "F#m",
      "body": "**Intro (4 bars)**\n\nF#m - A - E - B\n\n*Fingerpicking pattern on acoustic guitar*\n\n**Capo on 2nd fret**"
    },
    {
      "title": "Verse 1",
      "key": "F#m",
      "body": "**Verse 1**\n\nF#m          A           E           B\nToday is gonna be the day\n\nF#m          A           E           B\nThat they're gonna throw it back to you\n\n*Gentle strumming, focus on vocals*"
    },
    {
      "title": "Pre-Chorus",
      "key": "F#m",
      "body": "**Pre-Chorus**\n\nE           B           F#m\nBy now you should've somehow\n\nE           B           F#m\nRealized what you gotta do\n\n*Build up the energy*"
    },
    {
      "title": "Chorus",
      "key": "F#m",
      "body": "**Chorus**\n\nF#m          A           E           B\nI don't believe that anybody\n\nF#m          A           E           B\nFeels the way I do\n\nF#m          A           E           B\nAbout you now\n\n*Full strumming, strong vocals*"
    },
    {
      "title": "Verse 2",
      "key": "F#m",
      "body": "**Verse 2**\n\nF#m          A           E           B\nBackbeat, the word is on the street\n\nF#m          A           E           B\nThat the fire in your heart is out\n\n*Same pattern as verse 1*"
    },
    {
      "title": "Bridge",
      "key": "F#m",
      "body": "**Bridge**\n\nE           B           F#m\nAnd all the roads we have to walk are winding\n\nE           B           F#m\nAnd all the lights that lead us there are blinding\n\n*Emotional build-up*"
    },
    {
      "title": "Outro",
      "key": "F#m",
      "body": "**Outro**\n\nF#m - A - E - B (repeat)\n\n*Gradual fade out*\n\n**End on F#m**"
    }
  ]
}
```

Please generate song sections following this exact format and structure.
