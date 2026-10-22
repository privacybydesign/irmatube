package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed videos/*.json
var videoFiles embed.FS

type video struct {
	AgeRestriction int
	YoutubeID      string
}

// videoJSON is the file format. AgeRestriction is a pointer so that a
// missing or misspelled key is an error, not "no age limit".
type videoJSON struct {
	AgeRestriction *int   `json:"ageRestriction"`
	YoutubeID      string `json:"youtubeId"`
}

// videoCatalog maps a video id to its data. It is the allowlist of ids
// that the verify endpoint accepts.
type videoCatalog map[string]video

func loadVideos(fsys fs.FS) (videoCatalog, error) {
	names, err := fs.Glob(fsys, "videos/*.json")
	if err != nil {
		return nil, err
	}

	catalog := make(videoCatalog, len(names))
	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}

		var raw videoJSON
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if raw.AgeRestriction == nil {
			return nil, fmt.Errorf("%s: ageRestriction is missing", name)
		}

		id := strings.TrimSuffix(path.Base(name), ".json")
		catalog[id] = video{AgeRestriction: *raw.AgeRestriction, YoutubeID: raw.YoutubeID}
	}

	return catalog, nil
}
