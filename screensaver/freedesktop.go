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

package screensaver

import (
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
)

type Screensaver struct {
	bus          *dbus.Conn
	screenSaver  dbus.BusObject
	name, reason string

	cookie uint32
}

func NewScreensaver(name, reason string) (*Screensaver, error) {
	bus, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil, err
	}
	err = bus.Auth(nil)
	if err != nil {
		bus.Close()
		return nil, err
	}
	err = bus.Hello()
	if err != nil {
		bus.Close()
		return nil, err
	}
	return &Screensaver{
		bus: bus,
		screenSaver: bus.Object("org.freedesktop.ScreenSaver",
			"/org/freedesktop/ScreenSaver"),
		name:   name,
		reason: reason,
	}, nil
}

func (s *Screensaver) Inhibit() error {
	if s.cookie != 0 {
		return errors.New("Screensaver already inhibited")
	}
	var cookie uint32
	if err := s.screenSaver.Call("org.freedesktop.ScreenSaver.Inhibit", 0, s.name, s.reason).Store(&cookie); err != nil {
		return err
	}
	if cookie == 0 {
		return fmt.Errorf("invalid cookie (%d) received", s.cookie)
	}
	s.cookie = cookie
	return nil
}

func (s *Screensaver) Uninhibit() error {
	if s.cookie == 0 {
		return errors.New("Screensaver not inhibited")
	}
	if err := s.screenSaver.Call("org.freedesktop.ScreenSaver.UnInhibit", 0, s.cookie).Store(); err != nil {
		return err
	}
	s.cookie = 0
	return nil
}

func (s *Screensaver) Close() error {
	return s.bus.Close()
}
