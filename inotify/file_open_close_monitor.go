/*
 *    Copyright (c) 2023 Unrud <unrud@outlook.com>
 *
 *    This file is part of joystick-monitor.
 *
 *    joystick-monitor is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    joystick-monitor is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with joystick-monitor.  If not, see <http://www.gnu.org/licenses/>.
 */

package inotify

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"
	"unsafe"
)

type EventType int

const (
	EventOpen EventType = iota
	EventClose
	EventOverflow
)

type Event struct {
	Event EventType
	Path  string
}

type FileOpenCloseMonitor struct {
	inotify   *os.File
	watchPath string

	e chan error
	E <-chan error
	c chan Event
	C <-chan Event
}

func NewFileOpenCloseMonitor(watchPath string) (*FileOpenCloseMonitor, error) {
	var inotify *os.File
	if inotifyFd, err := syscall.InotifyInit1(syscall.IN_CLOEXEC | syscall.IN_NONBLOCK); err != nil {
		return nil, fmt.Errorf("InotifyInit1: %w", err)
	} else {
		inotify = os.NewFile(uintptr(inotifyFd), fmt.Sprintf("inotify(%d)", inotifyFd))
	}
	if _, err := syscall.InotifyAddWatch(int(inotify.Fd()), watchPath, syscall.IN_OPEN|syscall.IN_CLOSE); err != nil {
		inotify.Close()
		return nil, fmt.Errorf("InotifyAddWatch %v: %w", watchPath, err)
	}
	chanC := make(chan Event)
	chanE := make(chan error)
	m := &FileOpenCloseMonitor{
		inotify:   inotify,
		watchPath: watchPath,

		c: chanC,
		C: chanC,
		e: chanE,
		E: chanE,
	}
	go m.task()
	return m, nil
}

func (m *FileOpenCloseMonitor) task() {
	var buf [4096]byte
	for {
		size, err := m.inotify.Read(buf[:])
		if err != nil {
			m.e <- err
			return
		}
		eventsData := buf[:size]
		for {
			if len(eventsData) < syscall.SizeofInotifyEvent {
				m.e <- fmt.Errorf("read %v: %w", m.inotify.Name(), io.ErrUnexpectedEOF)
				return
			}
			event := (*syscall.InotifyEvent)(unsafe.Pointer(&eventsData[0]))
			eventsData = eventsData[syscall.SizeofInotifyEvent:]
			if event.Len > uint32(len(eventsData)) {
				m.e <- fmt.Errorf("read %v: %w", m.inotify.Name(), io.ErrUnexpectedEOF)
				return
			}
			eventName, _, _ := strings.Cut(string(eventsData[:int(event.Len)]), "\x00")
			eventsData = eventsData[int(event.Len):]
			if event.Mask&syscall.IN_IGNORED != 0 {
				m.e <- fmt.Errorf("read %v: watch %v ignored", m.inotify.Name(), m.watchPath)
				return
			}
			if event.Mask&syscall.IN_Q_OVERFLOW != 0 {
				m.c <- Event{EventOverflow, ""}
			} else if event.Mask&syscall.IN_ISDIR == 0 {
				if event.Mask&syscall.IN_OPEN != 0 {
					m.c <- Event{EventOpen, path.Join(m.watchPath, eventName)}
				}
				if event.Mask&syscall.IN_CLOSE != 0 {
					m.c <- Event{EventClose, path.Join(m.watchPath, eventName)}
				}
			}
			if len(eventsData) == 0 {
				break
			}
		}
	}
}

func (m *FileOpenCloseMonitor) Close() error {
	return m.inotify.Close()
}
