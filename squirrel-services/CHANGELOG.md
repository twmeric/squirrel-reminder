# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2024-01-15

### Added
- M04 Insight Engine - User profiling service
- Life event detection (job change, move, unemployment, retirement)
- Commute pattern analysis with regularity scoring
- Batch user analysis API (up to 100 users)
- Prometheus metrics integration
- Kubernetes deployment manifests
- Docker Compose local development setup

### Changed
- M03: Optimized GetSpeed latency from 11.5ms to 8ms P99
- M03: Improved TiDB connection pooling
- Enhanced staypoint detection algorithm

### Fixed
- Fixed race condition in state machine transitions
- Fixed memory leak in trajectory caching

## [1.1.0] - 2024-01-01

### Added
- m03-trajectory service with gRPC API
- GPS batch processing with DBSCAN staypoint detection
- Kalman filter for speed smoothing
- KD-tree based nearest station lookup
- Docker multi-stage builds

### Changed
- Upgraded to Go 1.21
- Migrated from MySQL to TiDB for distributed SQL

## [1.0.0] - 2023-12-15

### Added
- Initial release
- Basic trajectory storage
- Simple speed calculation
- Metro station data ingestion
