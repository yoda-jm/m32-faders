# m32-faders (pathfinder) — parked

This repository is **parked**. It was an early pathfinder for
**[tabula](https://github.com/yoda-jm/tabula)**, where development continues.

## What it was

A small Go prototype for controlling the faders of a Midas M32 / Behringer X32 console:
a web UI with live updates over WebSockets, a TCP command server
(`LIST`, `GET`, `SET … LEVEL`, `SET … MUTE`), and a fader layout configured in JSON.
The faders were virtual — the OSC link to a real console never went past a placeholder.

## Where it went

tabula picks up the idea properly: a real OSC codec and X32/M32 parameter schema,
a console emulator (`tabula-emu`) to develop without a console, and a full web GUI for
the desk and for the band's phones.

## Where's the code?

The last state of this prototype is preserved at the
[`pathfinder-final`](https://github.com/yoda-jm/m32-faders/tree/pathfinder-final) tag:

```sh
git checkout pathfinder-final
```
