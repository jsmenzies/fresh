# Release Process Update

## Summary

Updated the release workflow to use the consolidated Homebrew tap with modern GoReleaser configuration.

## Changes

### GoReleaser Configuration
- Migrated from deprecated `brews` configuration to `homebrew_casks`
- Consolidated to a single Homebrew tap: `jsmenzies/homebrew-tap`
- Removed separate scoop bucket in favor of using the same tap
- Simplified release artifact distribution

### Workflow Updates  
- Updated GitHub Actions to use latest versions (actions/setup-go@v6, actions/checkout@v6)
- Added Renovate for automated dependency updates
- Configured automated release process via release-please

## Release Flow

To trigger a new release:

1. **Merge changes to `main` branch** with conventional commits:
   - Use `feat:` for new features (bumps minor version)
   - Use `fix:` for bug fixes (bumps patch version)  
   - Use `docs:` or `refactor:` for visible changes (appears in changelog)
   - Avoid `chore:`, `ci:`, `test:` for hidden/internal changes

2. **Release-please automatically creates a Release PR** when it detects user-facing commits on main

3. **Merge the Release PR** to trigger the actual release:
   - Creates GitHub release with changelog
   - GoReleaser builds binaries for all platforms
   - Publishes to Homebrew tap
   - Generates checksums and release notes

## Versioning

- Current version: 1.9.2
- Next version will be determined by commit types since last release
- Version is tracked in `release/.release-please-manifest.json`
