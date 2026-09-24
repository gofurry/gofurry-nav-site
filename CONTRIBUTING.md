# Contributing to GoFurry

Contributions from developers and community members are welcome.
Please follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Choose the right channel

| Purpose | Channel |
| --- | --- |
| Questions and general discussion | [Discussions](https://github.com/gofurry/gofurry-nav-site/discussions) |
| Ideas and proposals | [Discussions](https://github.com/gofurry/gofurry-nav-site/discussions/categories/ideas) |
| Architecture and RFCs | [Discussions](https://github.com/gofurry/gofurry-nav-site/discussions) |
| Reproducible bugs | [Issues](https://github.com/gofurry/gofurry-nav-site/issues) |
| Accepted work items | [Issues](https://github.com/gofurry/gofurry-nav-site/issues) |
| Code changes | Pull Requests |
| Security vulnerabilities | Private reporting through [SECURITY.md](SECURITY.md) |

Discuss significant product or architecture changes before implementation.
Once a proposal is accepted, track the agreed scope in an Issue and link the
Discussion. Link the Issue from the implementing PR; small fixes can go straight
to a PR. Include reproduction steps and expected behavior in bug reports.

## Discussion categories

The intended categories are Announcements, General, Ideas, Q&A, Show and tell,
Polls, and **🏗 RFC / Architecture**.
Use RFC / Architecture for architecture decisions, major refactors, public API
design, cross-service changes, and technology selection. Until that category is
created, use General with an `[RFC]` title prefix.
Polls gather feedback; their results are not authoritative engineering decisions.

## Before opening a PR

1. Read [README.md](README.md) or [README_en.md](README_en.md).
2. Read [AGENTS.md](AGENTS.md), applicable module guidance, and relevant contracts.
3. Keep changes focused and within the owning service/module.
4. Run relevant tests and checks from the [playbook](.agents/playbook.md) and
   module documentation; report the results and anything not verified.
5. Explain the problem, resulting behavior, and related Issue/Discussion in the PR.
6. Never submit credentials, secrets, real production configuration, or private
   infrastructure values. Use sanitized examples when reporting problems.

## Licensing

By submitting a contribution to this repository, you agree to license that
contribution under the BSD 3-Clause License. You retain copyright in your work.
See [LICENSE](LICENSE) and [LICENSING.md](LICENSING.md) for the terms and boundaries.
Identify third-party material and preserve its original license and provenance;
only submit material you have the right to contribute.
No CLA, DCO sign-off, mandatory commit signing, or copyright assignment is required.
