package jwt

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Watcher struct {
	fd     int
	events map[int]string
}

func NewWatcher() (*Watcher, error) {
	fd, err := unix.InotifyInit()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		fd:     fd,
		events: make(map[int]string),
	}, nil
}

func (w *Watcher) Close() error {
	return unix.Close(w.fd)
}

func (w *Watcher) AddDirectory(path string) error {
	const (
		IN_MODIFY = 0x00000002
		IN_CREATE = 0x00000100
		IN_DELETE = 0x00000200
	)

	wd, err := unix.InotifyAddWatch(w.fd, path, IN_MODIFY|IN_CREATE)
	if err != nil {
		return err
	}

	w.events[wd] = path
	return nil
}

func (w *Watcher) WatchDirectories(path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return w.AddDirectory(path)
		}
		return nil
	})
}

func (w *Watcher) Next() (string, string, string, string, error) {
	var buf [unix.SizeofInotifyEvent + unix.PathMax]byte
	_, err := unix.Read(w.fd, buf[:])
	if err != nil {
		return "", "", "", "", err
	}

	raw := (*unix.InotifyEvent)(unsafe.Pointer(&buf[0]))
	nameBuf := buf[unix.SizeofInotifyEvent:]

	// Truncate at the first null byte
	nullPos := bytes.IndexByte(nameBuf, 0)
	if nullPos != -1 {
		nameBuf = nameBuf[:nullPos]
	}
	name := string(nameBuf)

	dir := w.events[int(raw.Wd)]

	print("file change event")

	var action string
	var type_ string
	var size int64

	if raw.Mask&unix.IN_MODIFY != 0 {
		action = "MODIFIED"
		type_ = "FILE"
		size = 0
	} else if raw.Mask&unix.IN_CREATE != 0 {
		action = "CREATE"

		if isDir(filepath.Join(dir, name)) {
			w.AddDirectory(filepath.Join(dir, name))
			type_ = "DIR"
			size = 0
		} else {
			type_ = "FILE"
			fileInfo, err := os.Stat(filepath.Join(dir, name))
			if err != nil {
				return "", "", "", "", err
			}
			size = fileInfo.Size()
		}
	}

	return action, type_, filepath.Join(dir, name), strconv.FormatInt(size, 10), nil
}

func isDir(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		return false
	}
	return info.IsDir()
}
