# Security Policy

## Supported Versions

adguardhome-sync does not maintain long-term support branches. Security
fixes are only provided for the **latest released version**.

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| Older   | :x:                |

Please make sure you're running the latest release before reporting an
issue — check the [Releases page](https://github.com/bakito/adguardhome-sync/releases).

## Reporting a Vulnerability

If you discover a security vulnerability in adguardhome-sync, please
**do not open a public GitHub issue**.

Instead, report it privately using GitHub's
[private vulnerability reporting](https://github.com/bakito/adguardhome-sync/security/advisories/new)
feature (Security tab → "Report a vulnerability").

When reporting, please include:
- A description of the vulnerability and its potential impact
- Steps to reproduce (config snippets, environment variables, etc. — with
  credentials/secrets redacted)
- The version of adguardhome-sync affected
- Any relevant logs

Since adguardhome-sync handles credentials for AdGuard Home instances
(origin/replica URLs, usernames, passwords, or cookies), please take extra
care not to include real credentials in your report.

## What to Expect

This is a community-maintained open source project run in the
maintainer's spare time, so response times are best-effort:

- Acknowledgement of your report: typically within a few days
- We'll work with you to understand and confirm the issue
- A fix will be prioritized based on severity and released as soon as
  practical
- Credit will be given in the release notes, unless you prefer to remain
  anonymous

## Scope Notes

adguardhome-sync connects to and stores credentials for your AdGuard Home
instances. If you are configuring it with sensitive credentials, review
the [README](https://github.com/bakito/adguardhome-sync#readme) for
guidance on secure configuration (e.g. environment variables vs. plain
YAML, TLS verification settings).

Vulnerabilities in **AdGuard Home itself** (the upstream project) should
be reported to the [AdGuardTeam/AdGuardHome](https://github.com/AdguardTeam/AdGuardHome)
repository, not here.
