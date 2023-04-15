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

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/unrud/joystick-monitor/inotify"
	"github.com/unrud/joystick-monitor/joystick"
	"github.com/unrud/joystick-monitor/processes"
	"github.com/unrud/joystick-monitor/screensaver"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	appName = "joystick-monitor"
	version = "0.0.3"

	ignoreMarker      = "ignore-joystick"
	maxRescanInterval = time.Second
	inhibitTimeout    = 10 * time.Second
)

func checkFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func orFatal[T any](value T, err error) T {
	checkFatal(err)
	return value
}

func keys[K comparable, V any](m map[K]V) (keys []K) {
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

type JoystickMonitorProxy struct {
	dev, ino uint64
	monitor  *joystick.JoystickMonitor

	closed     bool
	closeMutex sync.Mutex
}

func TryNewJoystickMonitorProxy(path string, activity chan struct{}) *JoystickMonitorProxy {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		return nil
	}
	checkFatal(err)
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		log.Fatal(err)
	}
	sysStat := stat.Sys().(*syscall.Stat_t)
	proxy := &JoystickMonitorProxy{dev: sysStat.Dev, ino: sysStat.Ino}
	if joystick.IsLegacyJoystickPath(path) {
		proxy.monitor = joystick.NewLegacyJoystickMonitor(file)
	} else {
		proxy.monitor = joystick.NewEventJoystickMonitor(file)
	}
	go proxy.task(activity)
	return proxy
}

func (proxy *JoystickMonitorProxy) task(activity chan struct{}) {
	for {
		select {
		case <-proxy.monitor.C:
			activity <- struct{}{}
		case err := <-proxy.monitor.E:
			if errors.Is(err, os.ErrClosed) || errors.Is(err, syscall.ENODEV) {
				proxy.Close()
				break
			}
			checkFatal(err)
		}
	}
}

func (proxy *JoystickMonitorProxy) IsSame(path string) bool {
	proxy.closeMutex.Lock()
	defer proxy.closeMutex.Unlock()
	if proxy.closed {
		return false
	}
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		return false
	}
	checkFatal(err)
	sysStat := stat.Sys().(*syscall.Stat_t)
	return proxy.dev == sysStat.Dev && proxy.ino == sysStat.Ino
}

func (proxy *JoystickMonitorProxy) Close() {
	proxy.closeMutex.Lock()
	defer proxy.closeMutex.Unlock()
	proxy.closed = true
	err := proxy.monitor.Close()
	if errors.Is(err, os.ErrClosed) {
		return
	}
	checkFatal(err)
}

func main() {
	var showVersion, dieWithParent bool
	flag.BoolVar(&dieWithParent, "die-with-parent", false, "exit program when parent terminates")
	flag.BoolVar(&showVersion, "version", false, "show program's version number and exit")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %v:\n", appName)
		flag.PrintDefaults()
	}
	flag.Parse()
	if showVersion {
		fmt.Println(version)
		return
	}

	if dieWithParent {
		checkFatal(processes.PrctlSetPdeathsig(syscall.SIGTERM))
	}
	ignoreMarkerFile := orFatal(processes.CreateMarker(ignoreMarker))
	defer ignoreMarkerFile.Close()
	joystickMonitorProxies := make(map[string]*JoystickMonitorProxy)
	defer func() {
		for _, proxy := range joystickMonitorProxies {
			proxy.Close()
		}
	}()
	inputFileMonitor := orFatal(inotify.NewFileOpenCloseMonitor("/dev/input"))
	defer inputFileMonitor.Close()
	screensaver := orFatal(screensaver.NewScreensaver(appName, "user activity"))
	defer screensaver.Close()
	rescanTimer := time.NewTimer(0)
	rescanTimerSet := true
	uninhibitTimer := time.NewTimer(0)
	if !uninhibitTimer.Stop() {
		<-uninhibitTimer.C
	}
	uninhibitTimerSet := false
	userActivity := make(chan struct{})
	for {
		select {
		case event := <-inputFileMonitor.C:
			if rescanTimerSet {
				continue
			}
			switch event.Event {
			case inotify.EventOpen:
				if proxy, found := joystickMonitorProxies[event.Path]; found {
					if proxy.IsSame(event.Path) {
						continue
					}
				} else if !joystick.IsLegacyJoystickPath(event.Path) {
					if _, found := orFatal(joystick.ListEventJoysticks())[event.Path]; !found {
						continue
					}
				}
			case inotify.EventClose:
				if _, found := joystickMonitorProxies[event.Path]; !found {
					continue
				}
			}
			rescanTimer.Reset(maxRescanInterval)
			rescanTimerSet = true
		case err := <-inputFileMonitor.E:
			checkFatal(err)
		case <-userActivity:
			if uninhibitTimerSet {
				if !uninhibitTimer.Stop() {
					<-uninhibitTimer.C
				}
			} else {
				checkFatal(screensaver.Inhibit())
				log.Println("inhibit")
			}
			uninhibitTimer.Reset(inhibitTimeout)
			uninhibitTimerSet = true
		case <-uninhibitTimer.C:
			uninhibitTimerSet = false
			checkFatal(screensaver.Uninhibit())
			log.Println("uninhibit")
		case <-rescanTimer.C:
			rescanTimerSet = false
			openJoystickPaths := orFatal(processes.FindOpenFiles(orFatal(joystick.ListAllJoysticks()), ignoreMarker))
			for path, proxy := range joystickMonitorProxies {
				if _, found := openJoystickPaths[path]; !found || !proxy.IsSame(path) {
					proxy.Close()
					delete(joystickMonitorProxies, path)
				}
			}
			for path := range openJoystickPaths {
				if _, found := joystickMonitorProxies[path]; !found {
					if monitor := TryNewJoystickMonitorProxy(path, userActivity); monitor != nil {
						joystickMonitorProxies[path] = monitor
					}
				}
			}
			log.Printf("scan [%v]\n", strings.Join(keys(joystickMonitorProxies), " "))
		}
	}
}
