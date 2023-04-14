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
	"fmt"
	"io"
	"math"
	"os"
	"unsafe"
)

type jsEvent struct {
	Time   uint32
	Value  int16
	Type   uint8
	Number uint8
}

const (
	jsEventButton = 0x01
	jsEventAxis   = 0x02
	jsEventInit   = 0x80
)

type legacyJoystickAxis struct {
	min, max int16
}

type legacyJoystickMonitor struct {
	JoystickMonitor
	axis map[uint8]legacyJoystickAxis
}

func NewLegacyJoystickMonitor(joystick *os.File) *JoystickMonitor {
	chanC := make(chan struct{})
	chanE := make(chan error)
	m := &legacyJoystickMonitor{
		JoystickMonitor: JoystickMonitor{joystick, chanC, chanC, chanE, chanE},
		axis:            make(map[uint8]legacyJoystickAxis),
	}
	go m.task()
	return &m.JoystickMonitor
}

func (m *legacyJoystickMonitor) task() {
	var buf [4096]byte
	for {
		size, err := m.joystick.Read(buf[:])
		if err != nil {
			m.e <- err
			return
		}
		eventsData := buf[:size]
		for {
			if len(eventsData) < int(unsafe.Sizeof(jsEvent{})) {
				m.e <- fmt.Errorf("read %v: %w", m.joystick.Name(), io.ErrUnexpectedEOF)
				return
			}
			event := (*jsEvent)(unsafe.Pointer(&eventsData[0]))
			eventsData = eventsData[int(unsafe.Sizeof(jsEvent{})):]
			if event.Type&jsEventAxis != 0 {
				state, stateSet := m.axis[event.Number]
				if !stateSet || event.Type&jsEventInit != 0 {
					state.min = event.Value
					state.max = event.Value
				} else {
					if event.Value < state.min {
						state.min = event.Value
					}
					if event.Value > state.max {
						state.max = event.Value
					}
					if uint16(state.max-state.min) > math.MaxUint16/8 {
						state.min = event.Value
						state.max = event.Value
						m.c <- struct{}{}
					}
				}
				m.axis[event.Number] = state
			}
			if event.Type == jsEventButton {
				m.c <- struct{}{}
			}
			if len(eventsData) == 0 {
				break
			}
		}
	}
}
