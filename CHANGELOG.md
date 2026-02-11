# Changelog

## [1.11.1](https://github.com/jsmenzies/fresh/compare/v1.11.0...v1.11.1) (2026-02-11)


### Performance Improvements

* parallelize BuildRepository and remove dead code ([e5adaf2](https://github.com/jsmenzies/fresh/commit/e5adaf26bfa7d74ea1a0533347ca810b85cb8219))
* parallelize BuildRepository and remove dead code ([793793e](https://github.com/jsmenzies/fresh/commit/793793ef939c6f81f5c899b62dc67114d58e9a94))
* parallelize BuildRepository and remove dead code ([90e5816](https://github.com/jsmenzies/fresh/commit/90e5816e9fca8204f796f22ca77945f83e75c39c))

## [1.11.0](https://github.com/jsmenzies/fresh/compare/v1.10.0...v1.11.0) (2026-02-11)


### Features

* dynamic column sizing for repo and branch names ([1e45d77](https://github.com/jsmenzies/fresh/commit/1e45d77cde65b167013f3bd7cce6f243e8e9f66c))
* dynamic column sizing for repo and branch names ([cd427fc](https://github.com/jsmenzies/fresh/commit/cd427fc7c21dc0728e1c610e09de4b94d259d5ee))
* dynamic column sizing for repo and branch names ([ffa391b](https://github.com/jsmenzies/fresh/commit/ffa391b05d06c229ba0e2f790d1f3a15795f74ff))


### Bug Fixes

* improve validateScanDir error handling ([ac83ed8](https://github.com/jsmenzies/fresh/commit/ac83ed8cedd9c9f9f8c246b773dbb04f20e7b0e9))


### Code Refactoring

* extract layout constants and pointer-based activities ([b7e321f](https://github.com/jsmenzies/fresh/commit/b7e321fe3a4aabed9489fb516bd561e5949ee550))
* extract layout constants from styles into dedicated layout module ([dd8521d](https://github.com/jsmenzies/fresh/commit/dd8521d70d9ab7d06e7009fbf814a55e3cd89a83))
* **git:** optimize FilterMergedBranches and add command timeouts ([5c6f2b5](https://github.com/jsmenzies/fresh/commit/5c6f2b5deb411f47222296af50d1ebabad1c5dd1))
* improve dynamic column sizing with centralized widths and helpers ([a3bfeba](https://github.com/jsmenzies/fresh/commit/a3bfebababdb00957162350ae42256fcff0bb077))
* improve dynamic column sizing with centralized widths and helpers ([5b2db66](https://github.com/jsmenzies/fresh/commit/5b2db66557c620062103f1fd4f5ef21dd8fbf328))
* move git operation timeouts to config ([c0770d7](https://github.com/jsmenzies/fresh/commit/c0770d79d7dbfc249947d765e68550e13b10c285))
* move git operation timeouts to config ([e25d78c](https://github.com/jsmenzies/fresh/commit/e25d78cad4c31f96b19e6c83f85d4bc4f5a38944))
* move git operation timeouts to config ([17afc89](https://github.com/jsmenzies/fresh/commit/17afc89bf5ef33bc33787f59a082eaba499b399f))
* move GitHub URL logic from UI to git package ([374ecb5](https://github.com/jsmenzies/fresh/commit/374ecb5dddf2027e98d80629e7a4acb2900dc707))
* use pointer-based activities to eliminate fragile copy-mutation pattern ([065d7ab](https://github.com/jsmenzies/fresh/commit/065d7ab9e790da8ed3d283a15b08c7d861c20fb8))
* use pointer-based activities to eliminate fragile copy-mutation pattern ([2ce3db2](https://github.com/jsmenzies/fresh/commit/2ce3db23d2793d957cb73ab4ab3530ca15d1ebbb))

## [1.10.0](https://github.com/jsmenzies/fresh/compare/v1.9.5...v1.10.0) (2026-02-11)


### Features

* add prune merged branches functionality ([ed15174](https://github.com/jsmenzies/fresh/commit/ed1517459f39c3c2f9a961e177cf6807bb309447))
* add prune merged branches functionality ([ea38d0b](https://github.com/jsmenzies/fresh/commit/ea38d0b238fd72df3f5c0ceaf321b895a4781bd6))
* consolidate keyboard shortcuts to single-key operations ([be641c3](https://github.com/jsmenzies/fresh/commit/be641c36f67cbc9d48f21dd51c427570d53ecfb1))
* consolidate keyboard shortcuts to single-key operations ([d8a6cde](https://github.com/jsmenzies/fresh/commit/d8a6cde6bf2dfea227c9f3e65fe64744aa0f21e1))
* separate regular merged and squashed branches with different keys ([75acc77](https://github.com/jsmenzies/fresh/commit/75acc77841c489957ed14ecb6db61635a009988c))
* separate regular merged and squashed branches with different keys ([8386906](https://github.com/jsmenzies/fresh/commit/83869063ade56414ee515362ec503b5626a9bdd0))
* show merged branches count in INFO column ([17d10be](https://github.com/jsmenzies/fresh/commit/17d10bef952ceb58638c41ea02fca90cf12e4b10))
* show merged branches count in INFO column ([a37987c](https://github.com/jsmenzies/fresh/commit/a37987cf68a66abaeb8b0385c1f5c1f08409c886))
* simplify branch pruning by removing squashed functionality ([75ea584](https://github.com/jsmenzies/fresh/commit/75ea584391a30e0a80a7561de591fd1e9bab098d))


### Bug Fixes

* **internal/git/git.go:** remove unused branch operations ([74b3503](https://github.com/jsmenzies/fresh/commit/74b35033d04d66b298bfb4d2b632904eeb28f521))


### Documentation

* **Release Process:** update release workflow to use Homebrew tap and GoReleaser configuration ([b9b392e](https://github.com/jsmenzies/fresh/commit/b9b392eef41f1c68c9b85432f3c983a4bb8b23e4))


### Code Refactoring

* add context timeouts to git commands and optimize branch filtering ([98d3819](https://github.com/jsmenzies/fresh/commit/98d38194cb69305eb43022feb10a062aabd9ca44))
* add context timeouts to git commands and optimize branch filtering ([bf3eebd](https://github.com/jsmenzies/fresh/commit/bf3eebdac650ea56eeb82cfe4adc1f27329b3e2e))
* group branch fields into Branches object ([a86103d](https://github.com/jsmenzies/fresh/commit/a86103dc09c15bef4d2e23763f90ed58fb8571e6))
* group branch fields into Branches object ([54b345a](https://github.com/jsmenzies/fresh/commit/54b345a4467db75ca086c248a3364a0ba584c590))
* remove unused code and extract shared LineBuffer ([f9c93b4](https://github.com/jsmenzies/fresh/commit/f9c93b477388c56fe7d7eb5bc4d575e0344b8d5d))

## [1.9.5](https://github.com/jsmenzies/fresh/compare/v1.9.4...v1.9.5) (2026-02-11)


### Bug Fixes

* **deps:** update module github.com/charmbracelet/bubbles to v1 ([2956881](https://github.com/jsmenzies/fresh/commit/295688123788ec502a433409d147a0334dd794bc))
* **deps:** update module github.com/charmbracelet/bubbles to v1 ([3437d1a](https://github.com/jsmenzies/fresh/commit/3437d1a38800c13813aae563af74b476ebd97f36))


### Documentation

* fix homebrew install command ([4506229](https://github.com/jsmenzies/fresh/commit/450622981a5c921a639e91fbd7d41c9efefb2144))
* update homebrew install command to use correct tap path ([3edb517](https://github.com/jsmenzies/fresh/commit/3edb5172bb9fb00c71fd346c8aa2e5e6a2a5b7bf))

## [1.9.4](https://github.com/jsmenzies/fresh/compare/v1.9.3...v1.9.4) (2026-02-03)


### Bug Fixes

* **deps:** update module github.com/charmbracelet/bubbles to v0.21.1 ([a9f6015](https://github.com/jsmenzies/fresh/commit/a9f601529e59b6f57abdd96cbcce0ad9d6dbdc37))
* **deps:** update module github.com/charmbracelet/bubbles to v0.21.1 ([3f15be6](https://github.com/jsmenzies/fresh/commit/3f15be6dac10bb32d28fbede11d118a730b3a9b7))

## [1.9.3](https://github.com/jsmenzies/fresh/compare/v1.9.2...v1.9.3) (2026-01-30)


### Documentation

* document release process changes and trigger release ([733d2be](https://github.com/jsmenzies/fresh/commit/733d2bed801337a235cb0f03b8de9ec4433e77b8))
* document release process migration to homebrew_casks ([b2cee92](https://github.com/jsmenzies/fresh/commit/b2cee92fe9e57ce5f483e0e34cae3e04d4b6a161))

## [1.9.2](https://github.com/jsmenzies/fresh/compare/v1.9.1...v1.9.2) (2026-01-26)


### Documentation

* update features and demo ([38cd766](https://github.com/jsmenzies/fresh/commit/38cd766f937793cb7f5df00b8996972973c6862f))
* update readme features and refresh demo gif ([628c2ba](https://github.com/jsmenzies/fresh/commit/628c2ba71bb41878b2dbbefef4bf42b5b9478a32))

## [1.9.1](https://github.com/jsmenzies/fresh/compare/v1.9.0...v1.9.1) (2026-01-03)


### Documentation

* clarify rebase usage and add no-icons to feature list ([402dd42](https://github.com/jsmenzies/fresh/commit/402dd42c4edeb673135b63e7f5946fe0a0f3a8b2))
* update README for alpha release ([0857e06](https://github.com/jsmenzies/fresh/commit/0857e06bdcfb6a75c333ba8f11d6241c07c211ee))


### Code Refactoring

* **ui:** simplify legend to a simple toggle and remove state-based filtering ([711db7e](https://github.com/jsmenzies/fresh/commit/711db7eed62d69113eb473a692d4e58ad833d4ce))
* **ui:** update legend rendering and adjust status column width ([73794c4](https://github.com/jsmenzies/fresh/commit/73794c413f4c27af5c2885d81f9c6da37c498db2))

## [1.9.0](https://github.com/jsmenzies/fresh/compare/v1.8.0...v1.9.0) (2026-01-03)


### Features

* **ui:** implement contextual status bar with help toggle ([a4ef2c3](https://github.com/jsmenzies/fresh/commit/a4ef2c387c42a3c6ab7029bae92f240f1f44e3db))


### Code Refactoring

* **ui:** move table view to listing package and implement legend grid ([680875b](https://github.com/jsmenzies/fresh/commit/680875bd29f367012384399a4c2e347dbdd772cf))

## [1.8.0](https://github.com/jsmenzies/fresh/compare/v1.7.3...v1.8.0) (2026-01-03)


### Features

* **ui:** highlight untracked files and add status legend ([dfc90b1](https://github.com/jsmenzies/fresh/commit/dfc90b192ffc10b2043eec3b80a58589b3bc0d57))
* **ui:** highlight untracked files and add status legend ([914f8e7](https://github.com/jsmenzies/fresh/commit/914f8e79822957e4308f9a2e680fc8d393b3b3dc))

## [1.7.3](https://github.com/jsmenzies/fresh/compare/v1.7.2...v1.7.3) (2025-12-27)


### Documentation

* add demo gif and tape ([43a6513](https://github.com/jsmenzies/fresh/commit/43a65136a49a5984c77ca128452c9b9fc1afc0f3))
* add demo gif and tape ([c9c0760](https://github.com/jsmenzies/fresh/commit/c9c0760d22358cd6c8c2e1963d2368487130fd9e))

## [1.7.2](https://github.com/jsmenzies/fresh/compare/v1.7.1...v1.7.2) (2025-12-27)


### Bug Fixes

* temporarily remove untracked icon and count, update readme ([2de8935](https://github.com/jsmenzies/fresh/commit/2de8935d2247c32023867e4060799f3ae6298847))
* temporarily remove untracked icon and count, update readme ([269c954](https://github.com/jsmenzies/fresh/commit/269c95452c0d4b5563cb08982aeac45c06fb95dd))

## [1.7.1](https://github.com/jsmenzies/fresh/compare/v1.7.0...v1.7.1) (2025-12-27)


### Documentation

* update readme usage and release description ([9e92487](https://github.com/jsmenzies/fresh/commit/9e924872df2e237d9f5996547ffc67594e5a4013))
* update readme usage and release description ([afb9772](https://github.com/jsmenzies/fresh/commit/afb977274f1a3610191f2e0df88e73d398bbbe59))

## [1.7.0](https://github.com/jsmenzies/fresh/compare/v1.6.1...v1.7.0) (2025-12-27)


### Features

* enhance pull logic and file status display ([794aa54](https://github.com/jsmenzies/fresh/commit/794aa5426051bc206af25b8b00f5d5090cb23133))
* enhance pull logic and file status display ([db792fd](https://github.com/jsmenzies/fresh/commit/db792fdd8cb2ce6656668ea9da0e9eef863c832b))

## [1.6.1](https://github.com/jsmenzies/fresh/compare/v1.6.0...v1.6.1) (2025-12-19)


### Code Refactoring

* use repo index for efficient lookup and style 'up to date' messages ([067f78d](https://github.com/jsmenzies/fresh/commit/067f78d52399e804a69c7ad6ba4e8cbb7a2cf135))
* use repo index for efficient lookup and style 'up to date' messages ([b9ca781](https://github.com/jsmenzies/fresh/commit/b9ca781c07568de956d142cb3773367c13b7ad91))

## [1.6.0](https://github.com/jsmenzies/fresh/compare/v1.5.0...v1.6.0) (2025-12-19)


### Features

* add version flag support (--version/-v) ([8cee4a4](https://github.com/jsmenzies/fresh/commit/8cee4a40ddb5ecc6d16f62df89440337ba9abe1a))
* **goreleaser:** add homebrew tap configuration ([83dc1ea](https://github.com/jsmenzies/fresh/commit/83dc1eae273c68b63fd96d8493633859133f2f1e))
* **listing:** modify keyboard shortcuts ([6f55b28](https://github.com/jsmenzies/fresh/commit/6f55b2842244d206a34134c508738cf20cf9d3d4))
* **ui:** Enhance remote status display and add simulation ([6592da1](https://github.com/jsmenzies/fresh/commit/6592da1a2bd8d514a6cfe919f1583a3d09a28f8b))
* **ui:** implement strictly truncated status messages ([9b5d08b](https://github.com/jsmenzies/fresh/commit/9b5d08b717e1e35ed1e84e1bdc691f87dbe581d3))


### Bug Fixes

* Configure release-please for main branch and changelog path ([3753b2b](https://github.com/jsmenzies/fresh/commit/3753b2b4783742b3a52ddbb94722f8e4a5f7ff81))
* Configure release-please to use main as primary branch and changelog path ([afbff6b](https://github.com/jsmenzies/fresh/commit/afbff6b62a1b86539791e4821d8060d8b7d25122))
* Correct GoReleaser config for proper archiving ([b247851](https://github.com/jsmenzies/fresh/commit/b2478518c81e38cb769cf0afcd54d67ba5dc1e80))
* Correct GoReleaser config for proper archiving ([8613a43](https://github.com/jsmenzies/fresh/commit/8613a438749d58656bfd1e58064d39ea6acf2c76))
* **goreleaser:** address archives format deprecation ([5208bd0](https://github.com/jsmenzies/fresh/commit/5208bd016533928ede155708e6948b4436842854))
* **goreleaser:** address archives format deprecation ([073a856](https://github.com/jsmenzies/fresh/commit/073a856e11f3bc2c987b0464906ea66f5b019519))
* **goreleaser:** correct ldflags format and use dynamic builtBy ([39e51dd](https://github.com/jsmenzies/fresh/commit/39e51ddb4b2ef66dc92e367a7262119f83253ca8))
* **goreleaser:** correct ldflags format and use dynamic builtBy ([804f81b](https://github.com/jsmenzies/fresh/commit/804f81be69c0c34df924f4d58c924ce993730886))
* **goreleaser:** Dynamically set builtBy in ldflags ([b300271](https://github.com/jsmenzies/fresh/commit/b3002713b0d2296af7e3736470e19c5c7f8fe92e))
* **goreleaser:** Dynamically set builtBy in ldflags ([8b381c9](https://github.com/jsmenzies/fresh/commit/8b381c9c93e569f3daf7c39758e11be31920b26a))
* **release:** correct tag format for goreleaser ([8e7bfaf](https://github.com/jsmenzies/fresh/commit/8e7bfafa5ddd7493abc41eb9cfa567a0f2db76c4))
* **release:** remove package-name from release-please config ([0a3a722](https://github.com/jsmenzies/fresh/commit/0a3a722530fea5f7e54045a1a7f952b153f48ef7))
* remove invalid package-name parameter from release workflow ([cfb5e97](https://github.com/jsmenzies/fresh/commit/cfb5e973cb5efbca92e2debb34fff74cf1babd37))
* Set Homebrew formula directory in GoReleaser ([39abffd](https://github.com/jsmenzies/fresh/commit/39abffd85e54f90515fbbaa0a5829b411f67b32d))
* Set Homebrew formula directory in GoReleaser ([788af04](https://github.com/jsmenzies/fresh/commit/788af04416b3926431774a10b50585dec01647b7))
* Update goreleaser to use homebrews ([f9510d0](https://github.com/jsmenzies/fresh/commit/f9510d083bb56d82a7b0672d7156de9f69674cec))
* Update goreleaser to use homebrews instead of brews ([05bacdd](https://github.com/jsmenzies/fresh/commit/05bacdd629ad14a795fd95ac026d1e3b582a0f11))
* updated release workflow permissions ([2853c9f](https://github.com/jsmenzies/fresh/commit/2853c9f7608de92f10029595f37421e052cc0ae3))


### Reverts

* Revert to using 'brews' in GoReleaser config ([84d8bfc](https://github.com/jsmenzies/fresh/commit/84d8bfc3a02c883f4620130cf7eef7ab2c204124))
* Revert to using 'brews' in GoReleaser config ([ba268df](https://github.com/jsmenzies/fresh/commit/ba268df6fc837d6d8f4252a3bd4b20554bb82aed))


### Documentation

* Update README description ([cbfbe63](https://github.com/jsmenzies/fresh/commit/cbfbe6311d62f13c8406c38fb17c04ff9be3f975))
* Update README description for multi-repo capabilities ([53a15ea](https://github.com/jsmenzies/fresh/commit/53a15ea726ac83c4249306b088b395206a0d43cc))


### Code Refactoring

* centralize info box width and fix layout stability ([d70b4db](https://github.com/jsmenzies/fresh/commit/d70b4db23e0040366ba0f29c90c03044291cd54c))
* centralize styles and formatting helpers ([09e2154](https://github.com/jsmenzies/fresh/commit/09e2154cece2ccddd277a13460866c44d1b52eaa))
* **formatting:** Move and update formatting package ([39ea572](https://github.com/jsmenzies/fresh/commit/39ea57232859348dd05329c029246fb55510aa01))
* state modelling for repositories and UI enhancements ([219e8ee](https://github.com/jsmenzies/fresh/commit/219e8ee45c5f82fb149d476b72c4d1d0ed001827))
* **ui:** consolidate style definitions and extract constants ([b1851fc](https://github.com/jsmenzies/fresh/commit/b1851fc001a42756bf9079f14b03a73bad6031e6))
* **ui:** Modularized UI views and fixed concurrency warning. Implemented pointer-based models for performance and to prevent lock value copying. ([aedd3e2](https://github.com/jsmenzies/fresh/commit/aedd3e23352b1a0b70a21dc1a972ea6fe280c0a1))
* **ui:** Modularized UI views and fixed concurrency warning. Implemented pointer-based models for performance and to prevent lock value copying. ([a197f8d](https://github.com/jsmenzies/fresh/commit/a197f8d5e0444368db15ae2398c6e9b518cf6d37))

## [1.5.0](https://github.com/jsmenzies/fresh/compare/v1.4.0...v1.5.0) (2025-12-19)


### Features

* **ui:** implement strictly truncated status messages ([9b5d08b](https://github.com/jsmenzies/fresh/commit/9b5d08b717e1e35ed1e84e1bdc691f87dbe581d3))


### Code Refactoring

* centralize info box width and fix layout stability ([d70b4db](https://github.com/jsmenzies/fresh/commit/d70b4db23e0040366ba0f29c90c03044291cd54c))
* centralize styles and formatting helpers ([09e2154](https://github.com/jsmenzies/fresh/commit/09e2154cece2ccddd277a13460866c44d1b52eaa))
* **ui:** consolidate style definitions and extract constants ([b1851fc](https://github.com/jsmenzies/fresh/commit/b1851fc001a42756bf9079f14b03a73bad6031e6))

## [1.4.0](https://github.com/jsmenzies/fresh/compare/v1.3.0...v1.4.0) (2025-12-17)


### Features

* **ui:** Enhance remote status display and add simulation ([6592da1](https://github.com/jsmenzies/fresh/commit/6592da1a2bd8d514a6cfe919f1583a3d09a28f8b))


### Code Refactoring

* **formatting:** Move and update formatting package ([39ea572](https://github.com/jsmenzies/fresh/commit/39ea57232859348dd05329c029246fb55510aa01))
* state modelling for repositories and UI enhancements ([219e8ee](https://github.com/jsmenzies/fresh/commit/219e8ee45c5f82fb149d476b72c4d1d0ed001827))

## [1.3.0](https://github.com/jsmenzies/fresh/compare/v1.2.7...v1.3.0) (2025-12-17)


### Features

* **listing:** modify keyboard shortcuts ([6f55b28](https://github.com/jsmenzies/fresh/commit/6f55b2842244d206a34134c508738cf20cf9d3d4))


### Bug Fixes

* **goreleaser:** correct ldflags format and use dynamic builtBy ([39e51dd](https://github.com/jsmenzies/fresh/commit/39e51ddb4b2ef66dc92e367a7262119f83253ca8))
* **goreleaser:** correct ldflags format and use dynamic builtBy ([804f81b](https://github.com/jsmenzies/fresh/commit/804f81be69c0c34df924f4d58c924ce993730886))

## [1.2.7](https://github.com/jsmenzies/fresh/compare/v1.2.6...v1.2.7) (2025-12-17)


### Bug Fixes

* **goreleaser:** Dynamically set builtBy in ldflags ([b300271](https://github.com/jsmenzies/fresh/commit/b3002713b0d2296af7e3736470e19c5c7f8fe92e))
* **goreleaser:** Dynamically set builtBy in ldflags ([8b381c9](https://github.com/jsmenzies/fresh/commit/8b381c9c93e569f3daf7c39758e11be31920b26a))

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
