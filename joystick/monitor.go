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
)

type JoystickMonitor struct {
	joystick *os.File

	c chan struct{}
	C <-chan struct{}
	e chan error
	E <-chan error
}

func (m *JoystickMonitor) Close() error {
	return m.joystick.Close()
}
