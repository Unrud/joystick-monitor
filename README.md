# Joystick Monitor

Monitors gamepads/joysticks used by applications and inhibits the screen saver during activity.

Supports the Linux
[event (evdev) interface](https://www.kernel.org/doc/html/v6.2/input/input.html#evdev) and
[legacy joystick API](https://www.kernel.org/doc/html/v6.2/input/joydev/joystick-api.html).
The screen saver is controlled with
[org.freedesktop.ScreenSaver](https://specifications.freedesktop.org/idle-inhibit-spec/latest/re01.html).

## Installation

### Fedora

```bash
sudo dnf copr enable unrud/joystick-monitor
sudo dnf install joystick-monitor
```

### Manual

```bash
# Install software
sudo wget -O /usr/local/bin/joystick-monitor \
  https://github.com/Unrud/joystick-monitor/releases/download/v0.0.3/joystick-monitor-linux-amd64
sudo chmod +x /usr/local/bin/joystick-monitor

# Create systemd service
sudo wget -O /etc/systemd/user/joystick-monitor.service \
  https://github.com/Unrud/joystick-monitor/releases/download/v0.0.3/joystick-monitor.service
```

## Setup

### Enable and start the service for the current user

```bash
systemctl --user enable --now joystick-monitor
```

### Enable service for all users

```bash
sudo systemctl --global enable joystick-monitor
```
