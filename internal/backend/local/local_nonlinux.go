//go:build !linux
// +build !linux

package local

import "os"

func readdirnames(f *os.File) ([]string, error) {
	return f.Readdirnames(-1)
}

func readdir(path string, f *os.File) ([]os.FileInfo, error) {
	return f.Readdir(-1)
}
