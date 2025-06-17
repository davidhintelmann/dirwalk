//go:build !windows
// +build !windows

package main

import (
	"io/fs"
)

func getFileSize(info fs.FileInfo) int64 {
	return info.Size()
}
