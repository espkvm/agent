# ESP-KVM agent

A small program for the machine [ESP-KVM](https://espkvm.io) is plugged into. It answers the device over the same USB cable that carries the keyboard and mouse, so the console can show what the machine is and share its clipboard. The network is not involved, and there is nothing to set up on the device side but one setting.

It is optional. ESP-KVM works without it.

## What it does

- Tells the device the machine's name, OS, uptime and IP addresses.
- Puts text on the clipboard, so the device does not have to type it.
- Reads the clipboard, only if you allow that on the machine, and never while the screen is locked.
- Shuts the machine down or restarts it cleanly, only if you allow that on the machine.

It does nothing on its own. It only answers what the device asks. [PROTOCOL.md](PROTOCOL.md) has the whole protocol.

## How it connects

ESP-KVM adds a USB serial port to the target when "Serial port over USB" in its settings is set to the agent. All three systems have the driver built in:

- Linux: /dev/ttyACM0
- Windows 10 and 11: a COM port
- macOS: /dev/cu.usbmodem…

The agent finds the port by its USB IDs (303A:4013), so the number it gets does not matter.

## Install

Download the file for your system from the releases.

Linux, as a system service (info and power, no clipboard). It runs as its own user `espkvm-agent`, not root:

```sh
sudo ./espkvm-agent-linux-amd64 install
```

Linux, in the desktop session (with the clipboard):

```sh
sudo ./espkvm-agent-linux-amd64 install --desktop
```

Both also add a udev rule that keeps ModemManager away from the port. It gives the port to the `espkvm-agent` user, or in desktop mode to the user at the screen.

Windows: run `espkvm-agent-windows-amd64.exe install` from a terminal. No administrator needed. It starts at every login.

macOS: `./espkvm-agent-darwin-arm64 install`. It starts at every login. The file is not signed yet, so remove the quarantine flag first: `xattr -d com.apple.quarantine espkvm-agent-darwin-arm64`.

Two switches add to what the device may do. Give them to `install` (or `run`):

- `--allow-clipboard-read`: the device may read the clipboard.
- `--allow-power`: the device may shut the machine down or restart it. For the Linux system service this also adds a polkit rule that lets the `espkvm-agent` user do that.

`uninstall` takes it all out again. The Linux `espkvm-agent` user stays.

To try it without installing, run it in a terminal: `espkvm-agent run`. `espkvm-agent info` prints what it would tell the device.

## Security

Whoever controls ESP-KVM already has this machine's keyboard and mouse, so the agent gives the device little it could not get by typing. It never runs commands. What it does add:

- It answers while the screen is locked, when typing gets nowhere. So it reads the clipboard only when you allow that, and not while the screen is locked. That check uses logind's LockedHint on Linux (GNOME, KDE and most lock screens set it), the input desktop on Windows, and the I/O registry on macOS.
- It tells the device the machine's name, OS and addresses, locked or not.
- It reads what comes over the USB cable. On Linux the system service runs as its own user with no capabilities and most of the system read-only, so a bug in the agent does not hand out root. On Windows and macOS it runs as the logged-in user.

The device alone can never switch on clipboard reading or power. Those switches are on the machine.

## Build

Go 1.26 or newer, no cgo. One machine builds every system:

```sh
sh build.sh
```

A tag like `v0.2.0` makes CI build every system and publish the release.

## License

Apache-2.0
