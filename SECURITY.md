# Security Policy

## Supported scope

| Branch or directory | Security support |
| --- | --- |
| `main` | Current production/release branch; security reports accepted. |
| `dev` | Active development branch; security reports accepted. |
| `legacy/**`, `experimental/**` | Not supported production software; no security maintenance guarantee. |
| `third-party/**` | Prefer reporting upstream; also report to GoFurry if its integration is directly affected. |

## Report privately

Do not publish unresolved vulnerabilities in public Issues or Discussions.
Email [2660621624@qq.com](mailto:2660621624@qq.com), the contact published in
[security.txt](apps/cn/nav-web/public/.well-known/security.txt).

Please include, when available:

- The affected component and version or commit.
- Reproduction steps and the conditions needed to trigger the issue.
- Expected impact and a proof of concept where appropriate.
- A way to contact you for follow-up questions.

Share only the information needed to explain the issue; redact credentials and
other people's private data. Coordinate public disclosure with the maintainer
through the private contact channel.

Reports are assessed according to impact and available maintenance capacity.
There is no guaranteed response or fix timeline, bug bounty, or monetary reward.
For ordinary reproducible bugs, use the workflow in [CONTRIBUTING.md](CONTRIBUTING.md).
