package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getVideos(mp4Videos *[]string, recursive bool) {
	var err error
	if (recursive) {
		err = filepath.WalkDir(".", func(path string, dir fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if (strings.HasSuffix(path, ".mp4") || strings.HasSuffix(path, ".mkv")) {
				*mp4Videos = append(*mp4Videos, path)
			} 
			return nil
		})
	} else {
		files, err := os.ReadDir(".")
		check(err)
		for _, directory := range files {
			if (strings.HasSuffix(directory.Name(), ".mp4") ||
				strings.HasSuffix(directory.Name(), ".mkv")) {
				*mp4Videos = append(*mp4Videos, directory.Name())
			}
		}
	}
	check(err)
}

func checkRecursive() bool {
	value := flag.Bool("recursive", false, "Recursively find video files")
	flag.Parse()
	return *value
}

func main() {
	videos := make([]string, 0)
	recursiveSearch := checkRecursive()

	getVideos(&videos, recursiveSearch)

	var totalTime time.Duration

	for _, file := range videos {
		cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1","-sexagesimal", file)
		durationBytes, err := cmd.Output() // duration is []byte 0:00:00.000000
		check(err)
		stringDuration := strings.Split(string(durationBytes), ".")[0]
		stringDuration = strings.Replace(stringDuration, ":", "h", 1)
		stringDuration = strings.Replace(stringDuration, ":", "m", 1)
		stringDuration = strings.Join([]string{stringDuration, "s"}, "")

		duration, err := time.ParseDuration(stringDuration)
		check(err)

		totalTime += duration
	}

	if (recursiveSearch) {
		fmt.Printf("Total duration (recursive): %v\n", totalTime)
	} else {
		fmt.Printf("Total duration: %v\n", totalTime)
	}
}
