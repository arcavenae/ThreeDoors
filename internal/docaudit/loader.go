package docaudit

import (
	"fmt"
	"path/filepath"
)

// LoadDocSet parses all four planning doc sources from the given project root.
func LoadDocSet(projectRoot string) (DocSet, error) {
	var ds DocSet
	var err error

	// 1. Story files.
	storyDir := filepath.Join(projectRoot, "docs", "stories")
	ds.StoryFiles, err = ParseStoryFiles(storyDir)
	if err != nil {
		return DocSet{}, fmt.Errorf("parse story files: %w", err)
	}

	// 2. Epic list.
	epicListPath := filepath.Join(projectRoot, "docs", "prd", "epic-list.md")
	ds.EpicList, err = ParseEpicList(epicListPath)
	if err != nil {
		return DocSet{}, fmt.Errorf("parse epic-list: %w", err)
	}
	ds.EpicListEpics = ds.EpicList

	// 3. Epics and stories.
	easPath := filepath.Join(projectRoot, "docs", "prd", "epics-and-stories.md")
	ds.EpicsAndStories, _, err = ParseEpicsAndStories(easPath)
	if err != nil {
		return DocSet{}, fmt.Errorf("parse epics-and-stories: %w", err)
	}

	// 4. ROADMAP.
	roadmapPath := filepath.Join(projectRoot, "ROADMAP.md")
	ds.Roadmap, ds.RoadmapEpics, err = ParseRoadmap(roadmapPath)
	if err != nil {
		return DocSet{}, fmt.Errorf("parse roadmap: %w", err)
	}

	return ds, nil
}
