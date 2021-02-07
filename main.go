package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Printf("%+v", err.Error())
	}
}

func run() error {
	dir := "./Takeout/Google Play Music/Tracks"
	log.Printf("reading %s", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	audio := make([]string, 0)
	tracks := make([]*track, 0)
	log.Printf("found %d files in %s", len(files), dir)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".csv") {
			audio = append(audio, file.Name())
			continue
		}

		song, err := parseCsv(filepath.Join(dir, file.Name()))
		if err != nil {
			log.Printf("failed to parse file: %s: %s", file.Name(), err.Error())
			continue
		}

		tracks = append(tracks, song)
	}

	resultPath := "result"
	if err := os.MkdirAll("result", os.ModePerm); err != nil {
		return fmt.Errorf("failed to write result directory")
	}

	for _, song := range audio {
		src := filepath.Join(dir, song)
		var dst string
		if Easy.MatchString(song) {
			matches := Easy.FindStringSubmatch(song)
			dst = filepath.Join(matches[1], matches[2], matches[3])
		} else if Hard.MatchString(song) {
			matches := Hard.FindStringSubmatch(song)
			dst = filepath.Join(matches[1], matches[2], matches[3])
		}

		if err := copyFile(src, filepath.Join(resultPath, dst)); err != nil {
			log.Printf("failed to copy file %s: %s", song, err.Error())
		}
	}

	return nil
}

var (
	Easy = regexp.MustCompile(`^(.*?) - (.*?) - (.*?\.mp3)$`)
	Hard = regexp.MustCompile(`^(.*?) - (.*?)\((\d{3}\).*?\.mp3)$`)
)

func parseCsv(path string) (*track, error) {
	fileName := filepath.Base(path)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed opening %s: %w", fileName, err)
	}

	data, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed reading csv data from %s: %w", fileName, err)
	}

	rawTrack := data[1]

	return &track{
		Title:      rawTrack[0],
		Album:      rawTrack[1],
		Artist:     rawTrack[2],
		DurationMs: rawTrack[3],
		Rating:     rawTrack[4],
		PlayCount:  rawTrack[5],
		Removed:    rawTrack[6],
	}, nil
}

type track struct {
	Title      string
	Album      string
	Artist     string
	DurationMs string
	Rating     string
	PlayCount  string
	Removed    string
}

func copyFile(src, dst string) error {
	audioFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write album: %w", err)
	}

	trackFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to write track file: %w", err)
	}

	if _, err := io.Copy(trackFile, audioFile); err != nil {
		return fmt.Errorf("failed to copyFile: %w", err)
	}
	_ = audioFile.Close()
	_ = trackFile.Close()
	return nil
}
