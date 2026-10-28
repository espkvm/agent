# ESP-KVM agent protocol

Version 1. The device and the agent talk over the USB serial port that ESP-KVM gives the target when "Serial port over USB" is set to the agent. The port is USB vendor 0x303A, product 0x4013.

## Lines

One JSON object per line, UTF-8, ending in `\n`. A `\r` before it is allowed. A line is at most 1 MiB. Longer lines are dropped whole.

The agent skips any line that does not start with `{` or is not a valid request. Whatever was on the port before the agent opened it does not break anything.

## Hello

When the agent opens the port it sends an empty line, then:

```json
{"ev":"hello","proto":1,"agent":"0.1.0","os":"linux","ops":["ping","info","clipboard.get","clipboard.set","power"]}
```

- `proto` is this version. A device that sees a higher number it does not know should still try the ops it knows.
- `os` is `linux`, `windows` or `darwin`.
- `ops` lists what this agent answers. An op that is switched off on the machine is still listed, it answers `not_allowed`.

The device can ask for the hello again with the `hello` op. That is also how it finds out whether an agent is there at all.

The hello is the only thing the agent sends unasked.

## Requests and responses

The device sends:

```json
{"id":7,"op":"clipboard.set","params":{"text":"hello"}}
```

The agent answers with the same `id`:

```json
{"id":7,"ok":true}
{"id":7,"ok":true,"data":{"text":"hello"}}
{"id":7,"ok":false,"error":"switched off on this machine","code":"not_allowed"}
```

The agent answers requests in the order they came. `error` is for a person, `code` is for the device:

| code | meaning |
|---|---|
| `unknown_op` | this agent does not have the op |
| `bad_params` | params missing or of the wrong type |
| `not_allowed` | the op is switched off on this machine |
| `no_desktop` | the clipboard needs the agent in a user's desktop session |
| `locked` | the screen is locked, so the clipboard is not read |
| `failed` | anything else, see `error` |

## Ops

### ping

No params. Answers `"pong"`.

### hello

No params. Answers the hello object.

### info

No params. Answers:

```json
{
  "hostname": "build-box",
  "os": "linux",
  "osName": "Ubuntu 24.04.4 LTS",
  "arch": "amd64",
  "uptimeS": 266420,
  "user": "anna",
  "ifaces": [{"name": "eth0", "mac": "f8:fe:5e:c3:71:ca", "ips": ["192.168.1.20", "fd00::20"]}]
}
```

`osName`, `uptimeS`, `user` and `mac` may be missing. `ifaces` lists interfaces that are up, not loopback, with an address that is not link-local.

### clipboard.get

No params. Answers `{"text": "..."}`. Text only, at most 512 KB.

Off unless the agent was started with `--allow-clipboard-read`. Even then it answers `locked` while the screen of its session is locked, and `failed` when it cannot tell whether it is.

### clipboard.set

Params `{"text": "..."}`, at most 512 KB. Answers with no data.

The clipboard belongs to a desktop session. An agent running as a system service (Linux without `--desktop`) answers `no_desktop`.

### power

Params `{"action": "shutdown"}` or `{"action": "reboot"}`. Answers `{"action": "..."}` first, then asks the OS to do it half a second later, the way the OS wants it done, so programs can close.

Off unless the agent was started with `--allow-power`. That switch is on the machine, not on the device, so the device alone can never turn it on.
