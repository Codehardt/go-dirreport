package dirreport

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

const (
	rootOld = "test_old/"
	rootNew = "test_new/"
)

func TestEmptyDirectories(t *testing.T) {
	defer reset()
	testDiff(t, nil, nil, nil, nil, nil)
}

func TestNewFile(t *testing.T) {
	defer reset()
	mustCreateFile(rootNew, "a", "")
	testDiff(t, []string{"a"}, nil, nil, nil, nil)
}

func TestRemovedFile(t *testing.T) {
	defer reset()
	mustCreateFile(rootOld, "a", "")
	testDiff(t, nil, []string{"a"}, nil, nil, nil)
}

func TestModifiedFile(t *testing.T) {
	defer reset()
	mustCreateFile(rootOld, "a", "1")
	mustCreateFile(rootNew, "a", "2")
	testDiff(t, nil, nil, []string{"a"}, nil, nil)
}

func TestNewDirectory(t *testing.T) {
	defer reset()
	mustCreateDirectory(rootNew, "a")
	testDiff(t, nil, nil, nil, []string{"a"}, nil)
}

func TestRemovedDirectory(t *testing.T) {
	defer reset()
	mustCreateDirectory(rootOld, "a")
	testDiff(t, nil, nil, nil, nil, []string{"a"})
}

func TestUnchangedFile(t *testing.T) {
	defer reset()
	mustCreateFile(rootOld, "a", "1")
	mustCreateFile(rootNew, "a", "1")
	testDiff(t, nil, nil, nil, nil, nil)
}

func TestUnchangedDir(t *testing.T) {
	defer reset()
	mustCreateDirectory(rootOld, "a")
	mustCreateDirectory(rootNew, "a")
	testDiff(t, nil, nil, nil, nil, nil)
}

func ExampleDirectoryReport_Diff() {
	defer reset()
	// before state
	mustCreateFile(rootOld, "file1.txt", "content a")
	mustCreateFile(rootOld, "file2.txt", "content b")
	before, _ := NewDirectoryReport(rootOld)
	// after state
	mustCreateFile(rootNew, "file1.txt", "content c") // changed content of file1.txt
	mustCreateFile(rootNew, "file3.txt", "content b") // renamed file2.txt to file3.txt
	after, _ := NewDirectoryReport(rootNew)
	newFiles, removedFiles, modifiedFiles, _, _ := before.Diff(after)
	fmt.Printf("New Files: %+v\n", newFiles)
	fmt.Printf("Removed Files: %+v\n", removedFiles)
	fmt.Printf("Modified Files: %+v\n", modifiedFiles)
	// Output:
	// New Files: [file3.txt]
	// Removed Files: [file2.txt]
	// Modified Files: [file1.txt]
}

func ExampleDirectoryReport_Diff_second() {
	defer reset()
	// before state
	mustCreateDirectory(rootOld, "dir1")
	mustCreateDirectory(rootOld, "dir1/subdir1")
	mustCreateDirectory(rootOld, "dir1/subdir2")
	before, _ := NewDirectoryReport(rootOld)
	// after state
	mustCreateDirectory(rootNew, "dir1")
	mustCreateDirectory(rootNew, "dir1/subdir1")
	mustCreateDirectory(rootNew, "dir1/subdir3") // renamed subdir2 to subdir 3
	after, _ := NewDirectoryReport(rootNew)
	_, _, _, newDirectories, removedDirectories := before.Diff(after)
	fmt.Printf("New Directories: %+v\n", newDirectories)
	fmt.Printf("Removed Directies: %+v\n", removedDirectories)
	// Output:
	// New Directories: [dir1/subdir3]
	// Removed Directies: [dir1/subdir2]
}

func ExampleDirectoryReport_Diff_third() {
	defer reset()
	// before state
	mustCreateDirectory(rootOld, "dir1")
	mustCreateDirectory(rootOld, "dir1/subdir1")
	mustCreateDirectory(rootOld, "dir1/subdir2")
	mustCreateFile(rootOld, "file1.txt", "content a")
	mustCreateFile(rootOld, "dir1/file2.txt", "content b")
	mustCreateFile(rootOld, "dir1/subdir1/file3.txt", "content c")
	mustCreateFile(rootOld, "dir1/subdir2/file4.txt", "content d")
	before, err := NewDirectoryReport(rootOld)
	if err != nil {
		panic(err)
	}
	// after state
	mustCreateDirectory(rootNew, "dir1")
	mustCreateDirectory(rootNew, "dir1/subdir1")
	/* mustCreateDirectory(rootNew, "dir1/subdir2") */ // this directory is removed
	mustCreateDirectory(rootNew, "dir1/subdir3")       // this directory is new
	mustCreateFile(rootNew, "file1.txt", "content a")
	mustCreateFile(rootNew, "dir1/file2.txt", "content b")
	mustCreateFile(rootNew, "dir1/subdir1/file3.txt", "modcontent c") // the content of this file has changed
	mustCreateFile(rootNew, "dir1/subdir3/file4.txt", "content d")    // this files has moved from subdir2 to subdir3
	mustCreateFile(rootNew, "dir1/subdir3/file5.txt", "content e")    // this files is new
	after, err := NewDirectoryReport(rootNew)
	if err != nil {
		panic(err)
	}
	newFiles, removedFiles, modifiedFiles, newDirectories, removedDirectories := before.Diff(after)
	fmt.Printf("New Files: %+v\n", newFiles)
	fmt.Printf("Removed Files: %+v\n", removedFiles)
	fmt.Printf("Modified Files: %+v\n", modifiedFiles)
	fmt.Printf("New Directories: %+v\n", newDirectories)
	fmt.Printf("Removed Directies: %+v\n", removedDirectories)
	// Output:
	// New Files: [dir1/subdir3/file4.txt dir1/subdir3/file5.txt]
	// Removed Files: [dir1/subdir2/file4.txt]
	// Modified Files: [dir1/subdir1/file3.txt]
	// New Directories: [dir1/subdir3]
	// Removed Directies: [dir1/subdir2]
}

func testDiff(t *testing.T, eNewFiles, eRemovedFiles, eModifiedFiles, eNewDirs, eRemovedDirs []string) {
	r1, err := NewDirectoryReport(rootOld)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := NewDirectoryReport(rootNew)
	if err != nil {
		t.Fatal(err)
	}
	for _, res := range [][]string{eNewFiles, eRemovedDirs, eModifiedFiles, eNewDirs, eRemovedDirs} {
		sort.Strings(res)
	}
	newFiles, removedFiles, modifiedFiles, newDirs, removedDirs := r1.Diff(r2)
	t.Logf("%s: diff: %+v, %+v, %+v, %+v, %+v", t.Name(), newFiles, removedFiles, modifiedFiles, newDirs, removedDirs)
	if !reflect.DeepEqual(newFiles, eNewFiles) {
		t.Fatalf("new files diff failed, expected: %+v, got: %+v", eNewFiles, newFiles)
	}
	if !reflect.DeepEqual(removedFiles, eRemovedFiles) {
		t.Fatalf("removed files diff failed, expected: %+v, got: %+v", eRemovedFiles, removedFiles)
	}
	if !reflect.DeepEqual(modifiedFiles, eModifiedFiles) {
		t.Fatalf("modified files diff failed, expected: %+v, got: %+v", eModifiedFiles, modifiedFiles)
	}
	if !reflect.DeepEqual(newDirs, eNewDirs) {
		t.Fatalf("new dirs diff failed, expected: %+v, got: %+v", eNewDirs, newDirs)
	}
	if !reflect.DeepEqual(removedDirs, eRemovedDirs) {
		t.Fatalf("removed dirs diff failed, expected: %+v, got: %+v", eRemovedDirs, removedDirs)
	}
}

func init() { reset() }

func reset() {
	if _, err := os.Stat(rootOld); err == nil || os.IsExist(err) {
		mustDeleteDirectory("", rootOld)
	} else if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	mustCreateDirectory("", rootOld)
	if _, err := os.Stat(rootNew); err == nil || os.IsExist(err) {
		mustDeleteDirectory("", rootNew)
	} else if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	mustCreateDirectory("", rootNew)
}

func mustCreateDirectory(root, path string) {
	if err := os.MkdirAll(filepath.Join(root, path), 0755); err != nil {
		panic(err)
	}
}

func mustDeleteDirectory(root, path string) {
	if err := os.RemoveAll(filepath.Join(root, path)); err != nil {
		panic(err)
	}
}

func mustCreateFile(root, path string, content string) {
	if err := ioutil.WriteFile(filepath.Join(root, path), []byte(content), 0644); err != nil {
		panic(err)
	}
}

func mustDeleteFile(root, path string) {
	if err := os.Remove(filepath.Join(root, path)); err != nil {
		panic(err)
	}
}
