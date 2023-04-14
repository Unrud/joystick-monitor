# Joystick Monitor

Monitors gamepads/joysticks used by applications and inhibits the screen saver during activity.

Supports the Linux
[event (evdev) interface](https://www.kernel.org/doc/html/v6.2/input/input.html#evdev) and
[legacy joystick API](https://www.kernel.org/doc/html/v6.2/input/joydev/joystick-api.html).
The screen saver is controlled with
[org.freedesktop.ScreenSaver](https://specifications.freedesktop.org/idle-inhibit-spec/latest/re01.html).

## Setup

Open a shell and copy/paste the following commands:

```bash
# Install software
sudo wget -O /usr/local/bin/joystick-monitor \
  https://github.com/Unrud/joystick-monitor/releases/download/v0.0.2/joystick-monitor-linux-amd64
sudo chmod +x /usr/local/bin/joystick-monitor

# Create systemd service
sudo wget -O /etc/systemd/user/joystick-monitor.service \
  https://github.com/Unrud/joystick-monitor/releases/download/v0.0.2/joystick-monitor.service

# Enable and start the service for the current user
systemctl --user enable --now joystick-monitor

# Alternative: Enable service for all users
sudo systemctl --global enable joystick-monitor
```
