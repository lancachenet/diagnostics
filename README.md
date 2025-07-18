# Lan Cache diagnostics utility

## Notes

This application has been created to assist in identifying any DNS misconfiguration on clients intending to use the Lan Cache stack.

## Usage

The application is a TUI (Terminal UI) which presents three modes when run.

* Diagnostics — Simple
* Diagnostics — Full
* Diagnostics — Custom

Executing any mode will write the output results to `diagnostics.txt` alongside the executable.

Below is an example of the output for a Diagnostics — Simple run:
```text
Interface: enp10s0
IP Address: 10.10.10.4/24
IP Address: 2400:a842:40bf:0:7d97:156c:564f:5741/64
IP Address: fe80::8669:d6ff:29a6:695a/64

DNS Server(s): 10.10.50.1, 10.10.10.254

Looking up Steam diagnostics address...
Successfully ran 6 diagnostics iteration(s) with system resolver
```

Diagnostics — Custom mode allows users to select which CDNs they would like to run the diagnostics tool against, this mode also allows filtering options by typing as demonstrated below for the Steam CDN:

[![asciicast](https://asciinema.org/a/728549.svg)](https://asciinema.org/a/728549)

Diagnostics — Full mode will run the diagnostics tool against all known CDNs as per the [cache-domains](https://github.com/uklans/cache-domains/) repository.
