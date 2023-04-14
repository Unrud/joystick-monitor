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
	"os"
	"path"
	"strconv"
	"strings"
)

func IsLegacyJoystickPath(path string) bool {
	const prefix = "/dev/input/js"
	if found := strings.HasPrefix(path, prefix); !found {
		return false
	}
	after := strings.TrimPrefix(path, prefix)
	nr, _ := strconv.Atoi(after)
	return strconv.Itoa(nr) == after
}

func ListLegacyJoysticks() (map[string]struct{}, error) {
	joysticks := make(map[string]struct{})
	inputDir, err := os.Open("/dev/input")
	if err != nil {
		return nil, err
	}
	defer inputDir.Close()
	inputEntries, err := inputDir.ReadDir(0)
	if err != nil {
		return nil, err
	}
	for _, inputEntry := range inputEntries {
		joystick := path.Join(inputDir.Name(), inputEntry.Name())
		if IsLegacyJoystickPath(joystick) {
			joysticks[joystick] = struct{}{}
		}
	}
	return joysticks, nil
}
