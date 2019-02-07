package dirreport

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// DebugFunc can be used to overwrite the Debugf function
type DebugFunc func(f string, a ...interface{})

// Debugf can be used to add a debug output writer
var Debugf = func(f string, a ...interface{}) {}

type DirectoryReport struct {
	root   string
	Files  map[string]string   `json:"files"`
	Dirs   map[string]struct{} `json:"dirs"`
	hasher hash.Hash
}

func NewDirectoryReport(path string) (*DirectoryReport, error) {
	directoryReport := &DirectoryReport{
		root:   path,
		Files:  map[string]string{},
		Dirs:   map[string]struct{}{},
		hasher: sha256.New(),
	}
	if err := filepath.Walk(path, directoryReport.walk); err != nil {
		Debugf("dirreport: could not walk through '%s': %s", path, err)
		return nil, err
	}
	return directoryReport, nil
}

func (r *DirectoryReport) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		Debugf("dirreport: handle dir '%s'", path)
		return r.handleDir(path, info)
	}
	if !info.Mode().IsRegular() {
		Debugf("dirreport: skipping non-regular file '%s'", path)
		return nil // skip special files
	}
	Debugf("dirreport: handle file '%s'", path)
	return r.handleFile(path, info)
}

func (r *DirectoryReport) handleDir(path string, _ os.FileInfo) error {
	relpath, err := filepath.Rel(r.root, path)
	if err != nil {
		Debugf("dirreport: could not get the relative path of dir '%s' in '%s': %s", path, r.root, err)
		return err
	}
	r.Dirs[relpath] = struct{}{}
	return nil
}

func (r *DirectoryReport) handleFile(path string, info os.FileInfo) error {
	relpath, err := filepath.Rel(r.root, path)
	if err != nil {
		Debugf("dirreport: could not get the relative path of file '%s' in '%s': %s", path, r.root, err)
		return err
	}
	r.hasher.Reset()
	f, err := os.Open(path)
	if err != nil {
		Debugf("dirreport: could not open file '%s': %s", path, err)
		return err
	}
	_, err = io.Copy(r.hasher, f)
	if errc := f.Close(); errc != nil {
		Debugf("dirreport: could not close file '%s': %s", path, errc)
		return errc
	}
	if err != nil {
		Debugf("dirreport: could not hash file '%s': %s", path, err)
		return err
	}
	r.Files[relpath] = hex.EncodeToString(r.hasher.Sum(nil))
	return nil
}

// Diff calculates the difference between two directory reports. It generates
// a list of new, removed and modified files and new dirs and removed dirs.
func (r *DirectoryReport) Diff(other *DirectoryReport) (newFiles, removedFiles, modifiedFiles, newDirs, removedDirs []string) {
	for relpath, otherSum := range other.Files {
		sum, ok := r.Files[relpath]
		if !ok {
			newFiles = append(newFiles, relpath)
		} else if sum != otherSum {
			modifiedFiles = append(modifiedFiles, relpath)
		}
	}
	for relpath := range r.Files {
		if _, ok := other.Files[relpath]; !ok {
			removedFiles = append(removedFiles, relpath)
		}
	}
	for relpath := range other.Dirs {
		if _, ok := r.Dirs[relpath]; !ok {
			newDirs = append(newDirs, relpath)
		}
	}
	for relpath := range r.Dirs {
		if _, ok := other.Dirs[relpath]; !ok {
			removedDirs = append(removedDirs, relpath)
		}
	}
	for _, res := range [][]string{newFiles, removedDirs, modifiedFiles, newDirs, removedDirs} {
		sort.Strings(res)
	}
	return
}
