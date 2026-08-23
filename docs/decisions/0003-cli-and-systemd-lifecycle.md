# 0003: CLI and systemd lifecycle

## Status

Accepted

## Context

Production services need a predictable foreground process and systemd registration that reflects the deployed binary, configuration, working directory, and runtime user.

## Decision

Every active Go binary uses Cobra. The root command displays help, `serve --config <file>` runs in the foreground, and Linux `install --config <file>` generates a systemd unit from the current executable, directory, user, and explicit configuration. Installation enables but never starts; `uninstall` removes only service registration.

## Alternatives Considered

Implicit root-command startup, static canonical production units, cross-platform service frameworks, and shell-generated command lines were rejected because they hide runtime inputs or produce environment-dependent units.

## Consequences

Operators review and start units explicitly. Deployment tooling must preserve foreground semantics and must not treat installation as an application restart.
