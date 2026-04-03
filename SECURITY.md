# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in redis-tui, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email **david@budnick.ca** with:

- A description of the vulnerability
- Steps to reproduce the issue
- Any potential impact

You can expect an initial response within 72 hours. Once the issue is confirmed, a fix will be prioritized and released as soon as possible.

## Scope

The following areas are in scope for security reports:

- Command injection or arbitrary code execution
- Credential exposure (Redis passwords, TLS keys)
- Path traversal in import/export operations
- Self-update mechanism integrity (checksum verification bypass)
- TLS/SSH configuration weaknesses

## Out of Scope

- Redis server-side vulnerabilities (report those to [Redis](https://github.com/redis/redis/security))
- Denial of service via terminal input
- Issues requiring physical access to the machine
