# livebuffer
A simple Go-based tool for snapshotting livestreams into segments and providing them via a REST API.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<br>

## Setup

### Docker (recommended)
The easiest way to run the tool is using Docker.  A pre-built image is available on the [GitHub Container Registry](https://github.com/matthiasharzer/livebuffer/pkgs/container/livebuffer).

#### Docker Compose
Create a `docker-compose.yml` file and start it with `docker compose up -d`. Make sure to adjust the command parameters as needed.

```yaml
services:
  livebuffer:
    image: ghcr.io/matthiasharzer/livebuffer:latest
    container_name: livebuffer
    restart: unless-stopped
    ports:
      - "4000:4000"

    command: run --port 4000 --interval 60 --history-size 24 --url https://www.youtube.com/watch?v=W0V8-6WrgBY
```
> [!NOTE]
> This example will snapshot the specified YouTube livestream every 60 minutes, keeping a history of the last 24 snapshots.

#### Docker CLI
```bash
docker run -d \
	--name livebuffer \
	-p 4000:4000 \
	ghcr.io/matthiasharzer/livebuffer:latest \
	run --port 4000 --interval 60 --history-size 24 --url https://www.youtube.com/watch?v=W0V8-6WrgBY
```

### Binary
Download the [latest release](https://github.com/matthiasharzer/livebuffer/releases/latest) for your platform and run it with the appropriate command-line arguments.

## Usage
Start the tool with:
```bash
./livebuffer run --port 4000 --interval 60 --history-size 24 --url https://www.youtube.com/watch?v=W0V8-6WrgBY
```

| Flag                | Required | Default               | Description                                                                                                                                                                                                                                      |
|---------------------|----------|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-u` / `--url`      | ✅       | /                     | The URL of the livestream to snapshot.                                                                                                                                                                                                           |
| `-p` / `--port`     | ❌       | 4000                  | The port on which the REST API will be available.                                                                                                                                                                                                |
| `--host`            | ❌       | `""` (all interfaces) | The host/IP address on which the REST API will listen.                                                                                                                                                                                           |
| `-i` / `--interval` | ❌       | 10                    | The interval (in minutes) at which to snapshot the livestream.                                                                                                                                                                                   |
| `--history-size`    | ❌       | 1                     | The number of snapshots to keep in history. Must be `>=1`                                                                                                                                                                                        |
| `--cookies-file`    | ❌       | `""`                  | Path to a cookies file (in Netscape format) for authenticated access. Will be passed as the `--cookies` flag to `yt-dlp`. See the [`yt-dlp` FAQ](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp) for further details. |

## API Endpoints
- `GET /api/v1/clip/{clip}`: Returns the n-th most recent snapshot, where `clip=0` is the most recent. Use instead `clip=latest` to always get the most recent snapshot. Returns the video clip in MP4 format.
