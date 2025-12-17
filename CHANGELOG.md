# Changelog

## [1.2.6](https://github.com/jsmenzies/fresh/compare/v1.2.5...v1.2.6) (2025-12-17)


### Code Refactoring

* **ui:** Modularized UI views and fixed concurrency warning. Implemented pointer-based models for performance and to prevent lock value copying. ([aedd3e2](https://github.com/jsmenzies/fresh/commit/aedd3e23352b1a0b70a21dc1a972ea6fe280c0a1))
* **ui:** Modularized UI views and fixed concurrency warning. Implemented pointer-based models for performance and to prevent lock value copying. ([a197f8d](https://github.com/jsmenzies/fresh/commit/a197f8d5e0444368db15ae2398c6e9b518cf6d37))

## [1.2.5](https://github.com/jsmenzies/fresh/compare/v1.2.4...v1.2.5) (2025-12-16)


### Bug Fixes

* Set Homebrew formula directory in GoReleaser ([39abffd](https://github.com/jsmenzies/fresh/commit/39abffd85e54f90515fbbaa0a5829b411f67b32d))
* Set Homebrew formula directory in GoReleaser ([788af04](https://github.com/jsmenzies/fresh/commit/788af04416b3926431774a10b50585dec01647b7))

## [1.2.4](https://github.com/jsmenzies/fresh/compare/v1.2.3...v1.2.4) (2025-12-16)


### Documentation

* Update README description ([cbfbe63](https://github.com/jsmenzies/fresh/commit/cbfbe6311d62f13c8406c38fb17c04ff9be3f975))
* Update README description for multi-repo capabilities ([53a15ea](https://github.com/jsmenzies/fresh/commit/53a15ea726ac83c4249306b088b395206a0d43cc))

## [1.2.3](https://github.com/jsmenzies/fresh/compare/v1.2.2...v1.2.3) (2025-12-16)


### Bug Fixes

* Configure release-please for main branch and changelog path ([3753b2b](https://github.com/jsmenzies/fresh/commit/3753b2b4783742b3a52ddbb94722f8e4a5f7ff81))
* Configure release-please to use main as primary branch and changelog path ([afbff6b](https://github.com/jsmenzies/fresh/commit/afbff6b62a1b86539791e4821d8060d8b7d25122))

## [1.2.2](https://github.com/jsmenzies/fresh/compare/v1.2.1...v1.2.2) (2025-12-16)


### Reverts

* Revert to using 'brews' in GoReleaser config ([84d8bfc](https://github.com/jsmenzies/fresh/commit/84d8bfc3a02c883f4620130cf7eef7ab2c204124))
* Revert to using 'brews' in GoReleaser config ([ba268df](https://github.com/jsmenzies/fresh/commit/ba268df6fc837d6d8f4252a3bd4b20554bb82aed))

## [1.2.1](https://github.com/jsmenzies/fresh/compare/v1.2.0...v1.2.1) (2025-12-15)


### Bug Fixes

* Update goreleaser to use homebrews ([f9510d0](https://github.com/jsmenzies/fresh/commit/f9510d083bb56d82a7b0672d7156de9f69674cec))
* Update goreleaser to use homebrews instead of brews ([05bacdd](https://github.com/jsmenzies/fresh/commit/05bacdd629ad14a795fd95ac026d1e3b582a0f11))

## [1.2.0](https://github.com/jsmenzies/fresh/compare/v1.1.1...v1.2.0) (2025-12-15)


### Features

* **goreleaser:** add homebrew tap configuration ([83dc1ea](https://github.com/jsmenzies/fresh/commit/83dc1eae273c68b63fd96d8493633859133f2f1e))


### Bug Fixes

* Correct GoReleaser config for proper archiving ([b247851](https://github.com/jsmenzies/fresh/commit/b2478518c81e38cb769cf0afcd54d67ba5dc1e80))
* Correct GoReleaser config for proper archiving ([8613a43](https://github.com/jsmenzies/fresh/commit/8613a438749d58656bfd1e58064d39ea6acf2c76))

## [1.1.1](https://github.com/jsmenzies/fresh/compare/v1.1.0...v1.1.1) (2025-12-15)


### Bug Fixes

* **goreleaser:** address archives format deprecation ([5208bd0](https://github.com/jsmenzies/fresh/commit/5208bd016533928ede155708e6948b4436842854))
* **goreleaser:** address archives format deprecation ([073a856](https://github.com/jsmenzies/fresh/commit/073a856e11f3bc2c987b0464906ea66f5b019519))

## [1.1.0](https://github.com/jsmenzies/fresh/compare/v1.0.0...v1.1.0) (2025-12-15)


### Features

* add version flag support (--version/-v) ([8cee4a4](https://github.com/jsmenzies/fresh/commit/8cee4a40ddb5ecc6d16f62df89440337ba9abe1a))


### Bug Fixes

* **release:** correct tag format for goreleaser ([8e7bfaf](https://github.com/jsmenzies/fresh/commit/8e7bfafa5ddd7493abc41eb9cfa567a0f2db76c4))
* **release:** remove package-name from release-please config ([0a3a722](https://github.com/jsmenzies/fresh/commit/0a3a722530fea5f7e54045a1a7f952b153f48ef7))
* remove invalid package-name parameter from release workflow ([cfb5e97](https://github.com/jsmenzies/fresh/commit/cfb5e973cb5efbca92e2debb34fff74cf1babd37))
* updated release workflow permissions ([2853c9f](https://github.com/jsmenzies/fresh/commit/2853c9f7608de92f10029595f37421e052cc0ae3))

## 1.0.0 (2025-12-15)


### Features

* add version flag support (--version/-v) ([8cee4a4](https://github.com/jsmenzies/fresh/commit/8cee4a40ddb5ecc6d16f62df89440337ba9abe1a))


### Bug Fixes

* remove invalid package-name parameter from release workflow ([cfb5e97](https://github.com/jsmenzies/fresh/commit/cfb5e973cb5efbca92e2debb34fff74cf1babd37))
* updated release workflow permissions ([2853c9f](https://github.com/jsmenzies/fresh/commit/2853c9f7608de92f10029595f37421e052cc0ae3))
