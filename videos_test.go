package main

import (
	"testing"
	"testing/fstest"
)

func TestLoadVideosNeedsAge(t *testing.T) {
	tests := map[string]string{
		"missing":    `{"youtubeId":"abc"}`,
		"misspelled": `{"ageRestricton":18,"youtubeId":"abc"}`,
	}

	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			fsys := fstest.MapFS{"videos/x.json": {Data: []byte(content)}}

			if _, err := loadVideos(fsys); err == nil {
				t.Errorf("loadVideos accepted a video without ageRestriction")
			}
		})
	}
}
