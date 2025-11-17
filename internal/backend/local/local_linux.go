package local

import (
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// start with a big buffer, so we probably won't have to realloc
const initialBufSize = 1024 * 1024

// use a smaller buffer to test if we need to restart
const continuationBufSize = 8192

// Get all directory entries in a single buffer.
// This works around a bug in how Go handles dirents with zero
// d_ino: https://github.com/golang/go/issues/76428
func getdentsOneShot(fd int) ([]byte, error) {
	contiuationBuf := make([]byte, continuationBufSize)
	bufSize := initialBufSize
	for {
		buf := make([]byte, bufSize)
		bytes, err := syscall.Getdents(fd, buf)
		if err != nil {
			return nil, err
		}

		// check if we got all the entries in our single call
		continuationBytes, err := syscall.Getdents(fd, contiuationBuf)
		if err != nil {
			return nil, err
		}
		if continuationBytes == 0 { // we got everything in one shot, good!
			return buf[:bytes], nil
		}

		// rewind and retry
		bufSize *= 2
		_, err = syscall.Seek(fd, 0, io.SeekStart)
		if err != nil {
			return nil, err
		}
	}
}

func readdirnames(f *os.File) ([]string, error) {
	buf, err := getdentsOneShot(int(f.Fd()))
	if err != nil {
		return nil, err
	}

	_, _, names := syscall.ParseDirent(buf, -1, []string{})
	return names, nil
}

func readdir(path string, f *os.File) ([]os.FileInfo, error) {
	names, err := readdirnames(f)
	if err != nil {
		return nil, err
	}
	infos := make([]os.FileInfo, 0, len(names))
	for _, name := range names {
		fpath := filepath.Join(path, name)
		info, err := os.Lstat(fpath)
		if os.IsNotExist(err) {
			continue // race condition, just ignore the file
		}
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}
