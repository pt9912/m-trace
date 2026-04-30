# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `0.1.0` Compose-Lab Core mit `api`, `mediamtx` und `stream-generator`.
- Root-Targets `make dev`, `make stop` und `make smoke`.
- MediaMTX-Konfiguration für RTSP-Publish, HLS auf Port `8888` und HTTP-API auf Port `9997`.
- FFmpeg-Teststream via `jrottenberg/ffmpeg:8.1-ubuntu2404`.

### Changed

- API-Listen-Adresse ist über `MTRACE_API_LISTEN_ADDR` konfigurierbar.
- Local-Development-Doku beschreibt den verifizierten `0.1.0`-Smoke-Test.

### Fixed
