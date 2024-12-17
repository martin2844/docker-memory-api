# Docker Volume Monitor

A simple HTTP service that monitors Docker container volumes and their usage statistics.

## Features

- Get volume statistics for all running containers
- Get volume statistics for a specific container by ID
- Returns container name, ID, volume name, usage (human-readable), usage in MB, and exposed port

## Prerequisites

- Go 1.23 or higher
- Docker installed and running
- Docker API version 1.43 or compatible

## Installation

1. Clone the repository
2. Install dependencies:
```
go mod download
```

## Docker Usage

To run as a Docker container:

```bash
# Build the image
docker build -t docker-image-api .

# Run the container with Docker socket access
docker run -d \
  --name docker-image-api \
  -p 6969:6969 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  docker-image-api
```

Note: The container requires access to the Docker socket (~/var/run/docker.sock~) to communicate with the Docker daemon and monitor container volumes. This gives the container privileged access to Docker operations on the host system.

## Usage

### Start the server:
```
go run main.go
```

The server will start on port 6969.

### API Endpoints

1. Get all container volume stats:
```
GET http://localhost:6969/stats
```

2. Get volume stats for a specific container:
```
GET http://localhost:6969/stats/{containerID}
```

### Example Response

```json
[
    {
        "container_name": "wp_dev-wordpress1-1",
        "container_id": "c8f7661f47fd",
        "volume_name": "a7bef70b2eb9c31a2163b3870db3f236219d9ead166eadd41490a6d601ac6e9a",
        "usage": "306.98MB",
        "usage_mb": "306.98",
        "port": "8080"
    }
]
```

## How it Works

The service uses the Docker API to:
1. List running containers
2. Inspect container details
3. Measure volume usage using a temporary busybox container
4. Extract port mappings from container network settings

## Dependencies

- github.com/docker/docker v27.4.0
- github.com/docker/go-connections v0.5.0
- github.com/docker/go-units v0.5.0

## License

MIT License

## Contributing

Feel free to open issues and pull requests for any improvements.