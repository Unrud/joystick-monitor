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

package joystick

import (
	"errors"
	"os"
	"path"
	"strings"
)

func ListEventJoysticks() (map[string]struct{}, error) {
	joysticks := make(map[string]struct{})
	inputByIdDir, err := os.Open("/dev/input/by-id")
	if errors.Is(err, os.ErrNotExist) {
		return joysticks, nil
	}
	if err != nil {
		return nil, err
	}
	defer inputByIdDir.Close()
	inputByIdEntries, err := inputByIdDir.ReadDir(0)
	if err != nil {
		return nil, err
	}
	for _, inputByIdEntry := range inputByIdEntries {
		if !strings.HasSuffix(inputByIdEntry.Name(), "-event-joystick") {
			continue
		}
		joystick, err := os.Readlink(path.Join(inputByIdDir.Name(), inputByIdEntry.Name()))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if path.IsAbs(joystick) {
			joystick = path.Clean(joystick)
		} else {
			joystick = path.Join(inputByIdDir.Name(), joystick)
		}
		joysticks[joystick] = struct{}{}
	}
	return joysticks, nil
}
