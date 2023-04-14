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

package processes

import (
	"errors"
	"os"
	"path"
	"strconv"
)

func FindOpenFiles(files map[string]struct{}, ignoreSelf bool, ignoreExe string) (openFiles map[string]struct{}, err error) {
	procDir, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer procDir.Close()
	procEntries, err := procDir.ReadDir(0)
	if err != nil {
		return nil, err
	}
	openFiles = make(map[string]struct{})
	if len(files) == 0 {
		return
	}
	for _, procEntry := range procEntries {
		pid, _ := strconv.Atoi(procEntry.Name())
		if strconv.Itoa(pid) != procEntry.Name() ||
			ignoreSelf && pid == os.Getpid() {
			continue
		}
		if err := func() error {
			if len(ignoreExe) != 0 {
				exe, err := os.Readlink(path.Join(procDir.Name(), procEntry.Name(), "exe"))
				if err != nil {
					return err
				}
				if exe == ignoreExe {
					return nil
				}
			}
			fdDir, err := os.Open(path.Join(procDir.Name(), procEntry.Name(), "fd"))
			if err != nil {
				return err
			}
			defer fdDir.Close()
			fdEntries, err := fdDir.ReadDir(0)
			if err != nil {
				return err
			}
			for _, fdEntry := range fdEntries {
				fdDest, err := os.Readlink(path.Join(fdDir.Name(), fdEntry.Name()))
				if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
					continue
				}
				if err != nil {
					return err
				}
				if _, found := files[fdDest]; found {
					openFiles[fdDest] = struct{}{}
				}
			}
			return nil
		}(); err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, os.ErrPermission) {
			return nil, err
		}
	}
	return
}
