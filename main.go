package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// VolumeStat holds the usage info for a container volume
type VolumeStat struct {
	ContainerName string `json:"container_name"`
	ContainerID   string `json:"container_id"`
	VolumeName    string `json:"volume_name"`
	Usage         string `json:"usage"`
	UsageMB       string `json:"usage_mb,omitempty"`
	Port          string `json:"port,omitempty"`
}

func main() {
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/stats/", singleContainerStatsHandler)
	fmt.Println("Listening on :6969...")
	log.Fatal(http.ListenAndServe(":6969", nil))
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.43"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []VolumeStat
	for _, c := range containers {

		inspect, err := cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			fmt.Printf("Failed to inspect container %s: %v\n", c.ID, err)
			continue
		}

		for _, mount := range inspect.Mounts {
			if mount.Type == "volume" {
				usageBytes := measureVolumeBytes(ctx, mount.Name)
				usageMB := fmt.Sprintf("%.2f", float64(usageBytes)/(1024*1024))

				// Get the first public port (if any)
				var port string
				for _, p := range inspect.NetworkSettings.Ports {
					if len(p) > 0 {
						port = fmt.Sprintf("%s", p[0].HostPort)
						break
					}
				}

				results = append(results, VolumeStat{
					ContainerName: strings.TrimPrefix(c.Names[0], "/"),
					ContainerID:   c.ID[:12],
					VolumeName:    mount.Name,
					Usage:         byteString(usageBytes), // e.g., "15M" from `du -sh`
					UsageMB:       usageMB,
					Port:          port,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func singleContainerStatsHandler(w http.ResponseWriter, r *http.Request) {
	containerID := strings.TrimPrefix(r.URL.Path, "/stats/")
	if containerID == "" {
		http.Error(w, "Container ID required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.43"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var results []VolumeStat
	for _, mount := range inspect.Mounts {
		if mount.Type == "volume" {
			usageBytes := measureVolumeBytes(ctx, mount.Name)
			usageMB := fmt.Sprintf("%.2f", float64(usageBytes)/(1024*1024))

			// Get the first public port (if any)
			var port string
			for _, p := range inspect.NetworkSettings.Ports {
				if len(p) > 0 {
					port = fmt.Sprintf("%s", p[0].HostPort)
					break
				}
			}

			results = append(results, VolumeStat{
				ContainerName: strings.TrimPrefix(inspect.Name, "/"),
				ContainerID:   containerID[:12],
				VolumeName:    mount.Name,
				Usage:         byteString(usageBytes),
				UsageMB:       usageMB,
				Port:          port,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// measureVolumeBytes returns the volume usage in bytes by running a
// short-lived busybox container that executes `du -sb /mnt`.
func measureVolumeBytes(ctx context.Context, volumeName string) int64 {
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", volumeName+":/mnt", "busybox", "du", "-sb", "/mnt")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error measuring volume %s: %v\n", volumeName, err)
		return 0
	}
	// Output should look like: "123456 /mnt"
	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		return 0
	}
	var size int64
	fmt.Sscanf(fields[0], "%d", &size)
	return size
}

// byteString runs a human-readable `du -sh` to get an approximate "usage" string (e.g. "15M")
func byteString(bytes int64) string {
	if bytes < 1<<10 {
		return fmt.Sprintf("%dB", bytes)
	} else if bytes < 1<<20 {
		return fmt.Sprintf("%.2fKB", float64(bytes)/1024)
	} else if bytes < 1<<30 {
		return fmt.Sprintf("%.2fMB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.2fGB", float64(bytes)/(1024*1024*1024))
}
