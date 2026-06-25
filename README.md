# tel42edgetts

tel42edgetts is an Asterisk AGI program designed to dynamically synthesize text-to-speech (TTS) using Microsoft Edge's high-quality neural voices.

## Features

- Microsoft Edge TTS API
- caching based on MD5 hashes of text + language + voice + format
- wav16 / wav / mp3 formats
- builtin mp3 to pcm convert
- **IVR mode**: user input can interrupt audio playback (single digit or digits ending with #)

## Installation

Run the build target via Make:

```bash
make build
```
This produces the statically linked executable binary `tel42edgetts`.

## Usage

You can call the executable out of an Asterisk dialplan via AGI, or run it directly from the command line for testing:

```bash
tel42edgetts [OPTIONS] [TEXT]
```

### Options

- `-lang`: Language for TTS (default: `en-US`).
- `-voice`: Voice name (default: `en-US-AvaMultilingualNeural`).
- `-format`: Output audio format (`mp3`, `wav`, `wav16`) (default: `wav16`).
- `-dir`: Directory to store cached audio files (default: `/tmp`).
- `-cache`: Enable caching of audio files. Use `-cache=false` to disable (default: `true`, env: `TTS_CACHE`).
- `-ivr`: Enable IVR mode - user input can interrupt playback (env: `TTS_IVR`).
- `-ivr-mode`: IVR input mode: `single` (single digit) or `hash` (digits ending with #) (env: `TTS_IVR_MODE`, default: `single`).
- `-ivr-timeout`: Timeout in milliseconds for waiting user input after playback completes (env: `TTS_IVR_TIMEOUT`, default: `5000`).
- `-version`: Print version and exit.

### Environment Variables

- `TTS_CACHE`: Set to `false` or `0` to disable caching by default. Can be overridden by the `-cache` command-line flag.
- `TTS_IVR`: Set to `true` or `1` to enable IVR mode by default.
- `TTS_IVR_MODE`: Set to `single` or `hash` to specify IVR input mode by default.
- `TTS_IVR_TIMEOUT`: Timeout in milliseconds for waiting user input after playback (default: `5000`).

### Asterisk Input Variables

You can override CLI flags by setting the following channel variables before calling the AGI:

- `TTS_TEXT`: The text you want to synthesize.
- `TTS_LANG`: Language code (e.g. `en-US`).
- `TTS_VOICE`: Target voice (e.g. `en-US-AvaMultilingualNeural`).
- `TTS_FORMAT`: Target format (e.g. `wav16`).
- `TTS_CACHE_DIR`: Directory for cache (e.g. `/tmp`).
- `TTS_CACHE`: Enable or disable caching (e.g. `false` or `0`).
- `TTS_IVR`: Enable IVR mode (e.g. `true` or `1`).
- `TTS_IVR_MODE`: IVR input mode (`single` or `hash`).
- `TTS_IVR_TIMEOUT`: Timeout in milliseconds for waiting user input after playback (default: `5000`).

### Asterisk Results Output

The script exports the outcome under the following Asterisk channel variables:

- `TTS_STATUS`: Will be set to `SUCCESS` if the audio was downloaded/processed successfully, or `ERROR` if the synthesis failed (e.g., missing text, network error).
- `TTS_USERINPUT`: Contains the user input received during IVR mode (empty string if no input).
  - In `single` mode: contains the single digit pressed (during or after playback)
  - In `hash` mode: contains the digit string ending with `#` (or empty if timeout)

### Asterisk Usage

```ini
...
same => n,Answer()
same => n,Set(TTS_TEXT=欢迎拨打智能语音服务，现在可以开始为您播放语音。)
same => n,Set(TTS_LANG=zh-CN)
same => n,Set(TTS_VOICE=Xiaoxiao)
same => n,Set(TTS_FORMAT=wav16)
same => n,Set(TTS_CACHE=true)

; Run the AGI script to synthesize and play the audio
same => n,AGI(tel42edgetts)

; Check the result
same => n,GotoIf($["${TTS_STATUS}"="SUCCESS"]?done:error)

; IVR Mode Example - single digit mode
same => n,Set(TTS_TEXT=Press 1 for sales, 2 for support.)
same => n,Set(TTS_IVR=true)
same => n,Set(TTS_IVR_MODE=single)
same => n,AGI(tel42edgetts)
same => n,Verbose(1,User pressed: ${TTS_USERINPUT})

; IVR Mode Example - hash mode
same => n,Set(TTS_TEXT=Please enter your extension number followed by hash.)
same => n,Set(TTS_IVR=true)
same => n,Set(TTS_IVR_MODE=hash)
same => n,AGI(tel42edgetts)
same => n,Verbose(1,User entered: ${TTS_USERINPUT})

; Fallback if TTS fails
same => n(error),Playback(vm-sorry)
same => n(done),Hangup()
...
```

## LICENSE

MIT
