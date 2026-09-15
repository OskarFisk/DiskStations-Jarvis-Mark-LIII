# DS413j package

This repository now contains a native DSM 6 package target for the Synology DS413j.

The DS413j uses Marvell Kirkwood 88F6282 and Synology package architecture `88f628x` (ARMv5). DSM 6.2.4 is the supported DSM family for this model.

The package is intentionally dependency-free at runtime: the service is a small statically-linked ARMv5 Go HTTP server. It does not require Docker, Container Manager, Python, Node.js, or third-party package repositories.

## Build

GitHub Actions builds `MARK-LIII-DS413j-1.0.0.spk` with a reproducible cross-build. The SPK contains:

- `INFO` with `arch="88f628x"`
- DSM lifecycle scripts
- ARMv5/GOARM=5 service binary
- static web dashboard

## Install

In DSM: Package Center -> Manual Install -> upload the generated `.spk`.

After installation, open:

`http://<NAS-IP>:8080/`

Configure the Gemini API key from the Settings section in the dashboard. The key is stored in the package data directory rather than the package itself.
