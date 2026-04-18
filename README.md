# livestream-snapshotting-tool
A simple Go-based tool for buffering livestreams and providing flexible access to stream clips via a REST API.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<br>

## Setup

### Docker (recommended)
The easiest way to run the tool is using Docker.  A pre-built image is available on the [GitHub Container Registry](https://github.com/matthiasharzer/livestream-snapshotting-tool/pkgs/container/livestream-snapshotting-tool).

#### Docker Compose
Create a `docker-compose.yml` file and start it with `docker compose up -d`. Make sure to adjust the command parameters as needed.

```yaml
services:
  livestream-snapshotting-tool:
    image: ghcr.io/matthiasharzer/livestream-snapshotting-tool:latest
    container_name: livestream-snapshotting-tool
    restart: unless-stopped
    ports:
      - "4000:4000"

    command: run --port 4000 --buffer 10m --url https://www.youtube.com/watch?v=W0V8-6WrgBY
```
> [!NOTE]
> This example will buffer the last 10 minutes of the livestream. Adjust the `--buffer` parameter as needed. See the [command-line flags](#usage) for more options.

#### Docker CLI
```bash
docker run -d \
	--name livestream-snapshotting-tool \
	-p 4000:4000 \
	ghcr.io/matthiasharzer/livestream-snapshotting-tool:latest \
	run --port 4000 --buffer 10m --url https://www.youtube.com/watch?v=W0V8-6WrgBY
```

### Binary
Download the [latest release](https://github.com/matthiasharzer/livestream-snapshotting-tool/releases/latest) for your platform and run it with the appropriate command-line arguments.

## Usage

### `run` Command
Start the tool with:
```bash
./livestream-snapshotting-tool run --port 4000 --buffer 10m --url https://www.youtube.com/watch?v=W0V8-6WrgBY
```

#### Command-Line Flags

| Flag              | Required | Default               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
|-------------------|----------|-----------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-u` / `--url`    | ✅       | /                     | The URL of the livestream to snapshot.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `-p` / `--port`   | ❌       | 4000                  | The port on which the REST API will be available.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `--host`          | ❌       | `""` (all interfaces) | The host/IP address on which the REST API will listen.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `-b` / `--buffer` | ❌       | 10m                   | The duration of the buffer, i.e., how much of the most recent livestream to keep available for snapshotting. Must be in a format parsable by [Go's `time.ParseDuration`](https://pkg.go.dev/time#ParseDuration) (e.g., `10m` for 10 minutes). Valid units are "ns", "us" (or "µs"), "ms", "s", "m", "h".                                                                                                                                                                                                                                                                                                                                              |
| `--buffer-dir`    | ❌       | _temporary directory_ | The directory where the livestream buffer will be stored. By default, a temporary directory will be created and used. If you want to persist the buffer across restarts, specify a directory here (e.g., `/data/buffer`).                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `--resume-buffer` | ❌       | false                 | Whether to attempt to resume the buffer from the specified `--buffer-dir` on startup. If enabled and a valid buffer is found in the directory, it will be loaded and used instead of starting with an empty buffer. This allows for continuity across restarts, but should only be used if you are sure that the buffer directory contains a valid and consistent buffer state. This may also create unwanted video jumps, since newly recorded video will be appended to the existing buffer, which may contain old video from a previous livestream session. If set to `false`, leftover buffer files will be deleted on startup. Use with caution. |
| `--cookies-file`  | ❌       | `""`                  | Path to a cookies file (in Netscape format) for authenticated access. Will be passed as the `--cookies` flag to `yt-dlp`. See the [`yt-dlp` FAQ](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp) for further details.                                                                                                                                                                                                                                                                                                                                                                                                      |

#### API Endpoints
| Method | Endpoint                               | Description                                                                                                                                                                                                                                                             |
|--------|----------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| GET    | `/api/v1/clip?start={start}&end={end}` | Request a clip of the livestream between the specified start and end times, in positive duration format.  For example, `start=5m&end=0s` would request a clip of the last 5 minutes of the livestream. The response will be a video file containing the requested clip. |

### `version` Command
Print the version of the tool:
```bash
./livestream-snapshotting-tool version
```

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details
