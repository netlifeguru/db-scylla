# Changelog

All notable changes to this project will be documented in this file.

This project follows:

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)

---

## [0.1.0] - 2026-05-23

### Added

- Initial public release
- ScyllaDB driver for the NetLife Guru Go database layer
- ScyllaDB database connection support
- Connection implementation compatible with the shared `db.Conn` interface
- Integration with `github.com/netlifeguru/db` query, exec, transaction, and repository helpers
- Mapper-backed result scanning through `github.com/netlifeguru/mapper`
- Struct, map, and scalar result handling through the shared database layer
- Support for Scylla `model.cql` files
- Scylla CQL dialect support through `db.DialectSQL`
- Batch workflow support through the shared database layer
- Documentation links for NetLife Guru docs and pkg.go.dev

### Notes

- This package is the ScyllaDB driver for the shared NetLife Guru database layer.
- Repository code can depend on `github.com/netlifeguru/db`, while application setup imports this driver.
- This is the first public `v0` release.
- The API may still change before `v1.0.0`.