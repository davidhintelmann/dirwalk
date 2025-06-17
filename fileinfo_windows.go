//go:build windows
// +build windows

package main

import (
	"io/fs"

	"golang.org/x/sys/windows"
)

// On Windows, info.Size() may not account for:
//
// - NTFS cluster rounding
//
// - Alternate data streams
//
// - Compression
//
// this function will format the filesize correctly on Win32 systems
func getFileSize(info fs.FileInfo) int64 {
	if stat, ok := info.Sys().(*windows.Win32FileAttributeData); ok {
		return int64(stat.FileSizeHigh)<<32 + int64(stat.FileSizeLow)
	}
	return info.Size()
}
