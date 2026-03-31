# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.1.1] - 2026-04-01

### Added

- Terminal resize support — UI adapts to window size changes
- Mouse interaction — click to select, double-click to switch profile, click buttons in help bar

### Fixed

- Removed unused formModel Y position tracking fields

## [0.1.0] - 2026-03-31

### Added

- Initial TUI implementation with Bubbletea framework
- Profile management: add, edit, delete configurations
- Profile switching with top-level field replacement
- Automatic backups before modifying settings.json (keep last 10)
- Secure storage with 0600 file permissions
- Cross-platform builds via goreleaser (macOS arm64/amd64, Linux amd64)
- GitHub Actions CI/CD for automated releases