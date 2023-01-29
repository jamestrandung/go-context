# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.4] - 2023-01-28
- Add generic type to memoize and cyclic features

## [1.0.3] - 2022-11-16
- Update memoize method signature to return more information with fewer outputs.
- Added a feature to pre-populate the cache for request-level memoization.
- Added a feature to find all memoized outcomes related to a particular execution key type.
- Added a feature to detect cyclic execution.

## [1.0.2] - 2022-11-11
- Use delegating context instead of detached context for memoization.

## [1.0.1] - 2022-11-10
- Added context for request-level memoization.

## [1.0.0] - 2022-05-05
### Added
- Initial release of this library as a Go module.
