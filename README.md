# stemsplit-cli

The official command-line interface for [StemSplit](https://stemsplit.io) — AI-powered audio stem separation.

## Install

### Homebrew (macOS / Linux)

```sh
brew tap StemSplit/tap
brew install stemsplit
```

### Direct download

Download the pre-built binary for your platform from [Releases](https://github.com/StemSplit/stemsplit-cli/releases).

### Build from source

```sh
git clone https://github.com/StemSplit/stemsplit-cli.git
cd stemsplit-cli
go build -o stemsplit .
```

## Quick start

```sh
# Set your API key (get one at https://stemsplit.io)
export STEMSPLIT_API_KEY=sk_live_xxxxxxxxxxxxx

# Or save it to config
stemsplit config set api_key sk_live_xxxxxxxxxxxxx

# Separate a song into vocals, drums, bass, and other
stemsplit separate song.mp3

# Extract only vocals
stemsplit separate song.mp3 --stems vocals

# Four stems, WAV output, custom output directory
stemsplit separate song.mp3 --stems vocals,drums,bass,other --format wav --output ./stems/
```

## Commands

### `stemsplit separate <file>`

Upload and separate an audio file into stems.

```
Flags:
  -s, --stems string         stems to extract (default "vocals,drums,bass,other")
                             values: vocals, instrumental, both, vocals/drums/bass/other, six-stems
  -m, --model string         model quality: htdemucs_ft (best), htdemucs (balanced), fast (default "htdemucs_ft")
  -o, --output string        output directory for downloaded stems (default ".")
      --format string        output format: mp3, wav, flac (default "mp3")
  -w, --wait                 wait for completion and auto-download stems (default true)
      --poll-interval int    seconds between status checks (default 3)
```

**Stem options:**

| `--stems` value | Description |
|---|---|
| `vocals` | Vocals only |
| `instrumental` | Instrumental (no vocals) |
| `both` | Vocals + instrumental (default 2-stem) |
| `vocals,drums,bass,other` | Four-stem separation |
| `vocals,drums,bass,other,piano,guitar` | Six-stem separation |

### `stemsplit jobs`

List recent jobs.

```
Flags:
  -n, --limit int   number of jobs to show (default 10, max 100)
```

### `stemsplit balance`

Show your credit balance. Displays a warning when balance is below 2 minutes.

### `stemsplit config set <key> <value>`

Set a configuration value (stored in `~/.config/stemsplit/config.json`).

```sh
stemsplit config set api_key sk_live_xxxxxxxxxxxxx
```

### `stemsplit config get <key>`

Get a configuration value.

```sh
stemsplit config get api_key
```

## Authentication

The API key is read in this order:

1. `STEMSPLIT_API_KEY` environment variable
2. `~/.config/stemsplit/config.json`

## Get an API key

Sign up at [stemsplit.io](https://stemsplit.io) — includes free minutes to get started.
