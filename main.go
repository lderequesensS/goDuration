package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type VideoType int

const (
	MP4 = iota
	MKV
)

type Video struct {
	name string
	videoType VideoType
}

func (v Video) GetDuration (b1 []byte, b3 []byte, b4 []byte, b8 []byte) (int64, error) {
	if (v.videoType == MP4) {
		file, err := os.Open(v.name)
		check(err)
		defer file.Close()
		for {
			read(file, b1)

			if rune(b1[0]) == 'm' {
				read(file, b3)

				if string(b3) == "vhd" {
					break
				}
			}
		}
		// We now have the correct index for "type"
		seek(file, -8, 1) // Move backward 8 bytes

		read(file, b4)
		size := int32(binary.BigEndian.Uint32(b4))
		check(err)
		if size == 0 || size == 1 {
			panic("So... size is 0 or 1 and... well I don't know what to do here")
		}

		seek(file, 4, 1)

		read(file, b1)
		version := int32(b1[0])

		seek(file, 3, 1)

		var timescale int32
		var duration int64

		if (version == 1) {
			seek(file, 16, 1)
			read(file, b4)
			timescale = int32(binary.BigEndian.Uint32(b4))
			read(file, b8)
			duration = int64(binary.BigEndian.Uint64(b8))
		} else {
			seek(file, 8, 1)
			read(file, b4)
			timescale = int32(binary.BigEndian.Uint32(b4))
			read(file, b4)
			duration = int64(binary.BigEndian.Uint32(b4))
		}

		return duration / int64(timescale), nil
	} else if (v.videoType == MKV) {
		return 0, nil // Not supported yet
	}
	return -1, errors.New("Format is not even thought to be supported")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func read(file *os.File, buffer []byte) {
	_, err := file.Read(buffer)
	check(err)
}

func seek(file *os.File, quantity int, mode int) {
	_, err := file.Seek(int64(quantity), mode)
	check(err)
}

func getVideos(mp4Videos *[]Video, mkvVideos *[]Video, recursive bool) {
	var err error
	if (recursive) {
		err = filepath.WalkDir(".", func(path string, dir fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if (strings.HasSuffix(path, ".mp4")) {
				*mp4Videos = append(*mp4Videos, Video{path, MP4})
			} else if (strings.HasSuffix(path, ".mkv")) {
				*mkvVideos = append(*mkvVideos, Video{path, MKV})
			}
			return nil
		})
	} else {
		files, err := os.ReadDir(".")
		check(err)
		for _, directory := range files {
			if (strings.HasSuffix(directory.Name(), ".mp4")) {
				*mp4Videos = append(*mp4Videos, Video{directory.Name(), MP4})
			} else if (strings.HasSuffix(directory.Name(), ".mkv")) {
				*mkvVideos = append(*mkvVideos, Video{directory.Name(), MKV})
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
	mp4Videos := make([]Video, 0)
	mkvVideos := make([]Video, 0)

	getVideos(&mp4Videos, &mkvVideos, checkRecursive())

	oneByteBuffer := make([]byte, 1)
	threeByteBuffer := make([]byte, 3)
	fourByteBuffer := make([]byte, 4)
	eightByteBuffer := make([]byte, 8)

	var mp4TotalSeconds int64 = 0
	var mkvTotalSeconds int64 = 0

	for _, filename := range mp4Videos {
		duration, err := filename.GetDuration(
			oneByteBuffer,
			threeByteBuffer,
			fourByteBuffer,
			eightByteBuffer,
		)
		check(err)

		mp4TotalSeconds += duration
	}

	for _, filename := range mkvVideos {
		duration, err := filename.GetDuration(
			oneByteBuffer,
			threeByteBuffer,
			fourByteBuffer,
			eightByteBuffer,
		)
		check(err)

		mkvTotalSeconds += duration
	}

	fmt.Println("MP4 files total time: ", time.Duration(mp4TotalSeconds) * time.Second)
	fmt.Println("MKV files total time: ", time.Duration(mkvTotalSeconds) * time.Second)
	fmt.Println("Total time: ", time.Duration(mp4TotalSeconds + mkvTotalSeconds) * time.Second)
}
