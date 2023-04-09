package yamlfmt

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/yamlfmt/internal/collections"
)

type PathCollector interface {
	CollectPaths() ([]string, error)
}

type FilepathCollector struct {
	Include    []string
	Exclude    []string
	Extensions []string
}

func (c *FilepathCollector) CollectPaths() ([]string, error) {
	fmt.Printf("[DEBUG ISSUE 97] entered filepath collector\n")
	pathsFound := []string{}
	for _, inclPath := range c.Include {
		info, err := os.Stat(inclPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		if !info.IsDir() {
			pathsFound = append(pathsFound, inclPath)
			continue
		}
		paths, err := c.walkDirectoryForYaml(inclPath)
		if err != nil {
			return nil, err
		}
		pathsFound = append(pathsFound, paths...)
	}

	pathsFoundSet := collections.SliceToSet(pathsFound)
	pathsToFormat := collections.SliceToSet(pathsFound)
	for _, exclPath := range c.Exclude {
		info, err := os.Stat(exclPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}

		absExclPath, err := filepath.Abs(exclPath)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			for foundPath := range pathsFoundSet {
				if strings.HasPrefix(foundPath, absExclPath) {
					pathsToFormat.Remove(foundPath)
				}
			}
		} else {
			pathsToFormat.Remove(absExclPath)
		}
	}

	return pathsToFormat.ToSlice(), nil
}

func (c *FilepathCollector) walkDirectoryForYaml(dir string) ([]string, error) {
	paths := []string{}
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		extension := ""
		if strings.Contains(info.Name(), ".") {
			nameParts := strings.Split(info.Name(), ".")
			extension = nameParts[len(nameParts)-1]
		}
		if collections.SliceContains(c.Extensions, extension) {
			paths = append(paths, path)
		}

		return nil
	})
	return paths, err
}

type DoublestarCollector struct {
	Include []string
	Exclude []string
}

func (c *DoublestarCollector) CollectPaths() ([]string, error) {
	fmt.Printf("[DEBUG ISSUE 97] entered doublestar collector\n")
	includedPaths := []string{}
	for _, pattern := range c.Include {
		globMatches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			fmt.Printf("[DEBUG ISSUE 97] filepath glob returned err=%#v\n", includedPaths)
			return nil, err
		}
		includedPaths = append(includedPaths, globMatches...)
	}
	fmt.Printf("[DEBUG ISSUE 97] includedPaths=%#v\n", includedPaths)

	pathsToFormatSet := collections.Set[string]{}
	for _, path := range includedPaths {
		if len(c.Exclude) == 0 {
			fmt.Printf("[DEBUG ISSUE 97] included because empty exclude path=%#v\n", path)
			pathsToFormatSet.Add(path)
			continue
		}
		excluded := false
		for _, pattern := range c.Exclude {
			absPath, err := filepath.Abs(path)
			if err != nil {
				// I wonder how this could ever happen...
				log.Printf("could not create absolute path for %s: %v", path, err)
				continue
			}
			match, err := doublestar.PathMatch(filepath.Clean(pattern), absPath)
			if err != nil {
				fmt.Printf("[DEBUG ISSUE 97] path match returned err=%#v\n", err)
				return nil, err
			}
			if match {
				fmt.Printf("[DEBUG ISSUE 97] excluded because of pattern=%#v absPath=%#v\n", pattern, absPath)
				excluded = true
			}
			fmt.Printf("[DEBUG ISSUE 97] no match pattern=%#v absPath=%#v\n", pattern, absPath)
		}
		if !excluded {
			fmt.Printf("[DEBUG ISSUE 97] included because not excluded path=%#v\n", path)
			pathsToFormatSet.Add(path)
		}
	}

	return pathsToFormatSet.ToSlice(), nil
}
