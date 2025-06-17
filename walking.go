package main

import (
	"container/heap"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// walk directories sequentially using `filepath.WalkFunc` from golang standard
// library. Not as well optimized as `ConcurrentWalk` function which uses concurrency
func ProcessDirectorySimple(root string, maxDepth int) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// handle PerfLogs Admin Access by skipping directory
			if info.IsDir() && info.Name() == "PerfLogs" {
				log.Println("Skipping PerfLogs...")
				return fs.SkipDir
			}
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == "$Recycle.Bin" {
			log.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		} else if info.IsDir() && info.Name() == "PerfLogs" {
			log.Println("skip PerfLogs", path)
			return fs.SkipDir
		} else if info.IsDir() && dirHelper(root, path) >= maxDepth {
			log.Printf("Reached max depth: %v, subfolder: %s\n", maxDepth, path)
			return fs.SkipDir
		} else {
			fsize, funit := formatSize(getFileSize(info))
			log.Printf("visited file or dir: %q -- filesize: %s %s\n", path, fsize, funit)
			return nil
		}
	}
}

// helper function to return depth of recursive walk as an int
// starting from the root directory
func dirHelper(root, subfolder string) int {
	after, found := strings.CutPrefix(subfolder, root)
	if !found {
		log.Fatalf("did not cut root prefix from path.\nroot:%s\nsubfolder:%s\nafter:%s\n", root, subfolder, after)
	}

	return strings.Count(after, string(os.PathSeparator))
}

// helper function to format the size of `fs.FileInfo` from
// bytes into KB, MB, or GB if necessary
func formatSize(filesize int64) (string, string) {
	fsize := float64(filesize)
	sizes := [4]string{"bytes", "KB", "MB", "GB"}
	var idx int
	for fsize > 1024 && idx < 3 {
		fsize /= 1024
		idx++
	}

	if sizes[idx] == "bytes" {
		result := fmt.Sprint(fsize)
		return result, sizes[idx]
	} else {
		result := fmt.Sprintf("%.2f", fsize)
		return result, sizes[idx]
	}
}

type StatsOutput struct {
	filesScanned int64
	dirsScanned  int64
	totalSize    int64
}

func (so *StatsOutput) PrintOutput() {
	p := message.NewPrinter(language.English)
	p.Printf("\nScanned %d directories\n", so.dirsScanned)
	p.Printf("Scanned %d files\n", so.filesScanned)
	fsize, funit := formatSize(so.totalSize)
	p.Printf("Total disk space scanned %s %s\n", fsize, funit)
}

type FileEntry struct {
	Path string
	Size int64
}

type FileMinHeap []FileEntry

func (h FileMinHeap) Len() int           { return len(h) }
func (h FileMinHeap) Less(i, j int) bool { return h[i].Size < h[j].Size }
func (h FileMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *FileMinHeap) Push(x any) {
	*h = append(*h, x.(FileEntry))
}

func (h *FileMinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// Optimized algorithm for walking directories starting at root. Uses
// concurrency to efficiently process directory stats with multiple
// go routines. Currently set to 64 parallel processes. Only to be used
// on Windows operating system. For Unix based system use the
// `ConcurrentWalkUnix` function instead.
func ConcurrentWalk(root string, maxDepth, topN int) *StatsOutput {
	results := &StatsOutput{}
	semaphore := make(chan struct{}, 64)
	var wg sync.WaitGroup
	var mu sync.Mutex

	h := &FileMinHeap{}
	heap.Init(h)

	semaphore <- struct{}{} // acquire root
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { <-semaphore }()
		walkDir(root, root, maxDepth, topN, results, &wg, h, &mu, semaphore)
	}()
	wg.Wait()

	fmt.Println("\nTop", topN, "largest files:")
	sorted := make([]FileEntry, h.Len())
	for i := len(sorted) - 1; i >= 0; i-- {
		sorted[i] = heap.Pop(h).(FileEntry)
	}

	for i, f := range sorted {
		size, unit := formatSize(f.Size)
		fmt.Printf("%2d. %-10s %s\n", i+1, size+" "+unit, f.Path)
	}
	return results
}

func walkDir(root, current string, maxDepth, topN int, results *StatsOutput, wg *sync.WaitGroup, h *FileMinHeap, mu *sync.Mutex, semaphore chan struct{}) {
	entries, err := os.ReadDir(current)
	if err != nil {
		atomic.AddInt64(&results.dirsScanned, 1) // increment directory counter
		// log.Printf("failed to read %s: %v\n", current, err)
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(current, entry.Name())
		info, err := entry.Info()
		if err != nil {
			log.Printf("could not get info for %s: %v\n", fullPath, err)
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if entry.IsDir() {
			atomic.AddInt64(&results.dirsScanned, 1) // increment directory counter
			depth := dirHelper(root, fullPath)
			if maxDepth >= 0 && depth > maxDepth {
				continue
			}

			select {
			case semaphore <- struct{}{}: // acquire before spawn
				wg.Add(1) // add before spawn
				go func(p string) {
					defer wg.Done()
					defer func() { <-semaphore }()
					walkDir(root, p, maxDepth, topN, results, wg, h, mu, semaphore)
				}(fullPath)
			default:
				// fallback to synchronous call when no slots available
				walkDir(root, fullPath, maxDepth, topN, results, wg, h, mu, semaphore)
			}

		} else {
			atomic.AddInt64(&results.filesScanned, 1)              // increment file counter
			atomic.AddInt64(&results.totalSize, getFileSize(info)) // add to disk space total
			mu.Lock()
			if h.Len() < topN {
				heap.Push(h, FileEntry{Path: fullPath, Size: getFileSize(info)})
			} else if (*h)[0].Size < getFileSize(info) {
				heap.Pop(h)
				heap.Push(h, FileEntry{Path: fullPath, Size: getFileSize(info)})
			}
			mu.Unlock()
		}
	}
}
