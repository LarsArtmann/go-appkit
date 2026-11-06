# Changelog

All notable changes to this module are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned

- Initial module (SUPERB v12 P-train): `preset.New(Config)` one-call
  composition — always-on spine (core Service, systemd.Install, health
  dashboard, error pages, RegisterClassifications, version stamp) plus
  env-driven legs (`APPKIT_OTEL`, `APPKIT_REALTIME`, `APPKIT_ERRORPAGES`,
  `APPKIT_SECURITY`) with explicit-code-wins precedence and fail-fast
  Rejection on invalid values.
