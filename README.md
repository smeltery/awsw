# awsw

![awsw](https://raw.githubusercontent.com/smeltery/awsw/master/assets/og-image.svg)

[![CI](https://github.com/smeltery/awsw/actions/workflows/ci.yml/badge.svg)](https://github.com/smeltery/awsw/actions/workflows/ci.yml)
[![Release](https://github.com/smeltery/awsw/actions/workflows/release.yml/badge.svg)](https://github.com/smeltery/awsw/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/smeltery/awsw)](https://goreportcard.com/report/github.com/smeltery/awsw)
[![License: PolyForm Shield 1.0.0](https://img.shields.io/badge/License-PolyForm%20Shield%201.0.0-blue.svg)](https://polyformproject.org/licenses/shield/1.0.0/)

![Go](https://img.shields.io/badge/-Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![AWS](https://img.shields.io/badge/-AWS_SSO-232F3E?style=flat-square&logo=amazonwebservices&logoColor=white)
![Kubernetes](https://img.shields.io/badge/-Kubernetes-326CE5?style=flat-square&logo=kubernetes&logoColor=white)
![Cobra](https://img.shields.io/badge/-Cobra-00ADD8?style=flat-square&logo=go&logoColor=white)
![macOS](https://img.shields.io/badge/-macOS-000000?style=flat-square&logo=apple&logoColor=white)
![Linux](https://img.shields.io/badge/-Linux-FCC624?style=flat-square&logo=linux&logoColor=black)

> *"One command, any account."*

A lightweight CLI for seamlessly switching between AWS SSO accounts and EKS cluster contexts. Switch account → switch kubectl context → use `k9s` / `kubectl` against the right cluster immediately.

## Installation

### Via Homebrew

```sh
brew install --cask smeltery/tap/awsw
```

### From source

```sh
git clone https://github.com/smeltery/awsw.git
cd awsw
make install
```

## Quickstart

```sh
# 1. Generate AWS config and discover EKS clusters
awsw setup

# 2. Authenticate via SSO (opens browser)
awsw login

# 3. Switch to an account
awsw dev-eks

# 4. Use kubectl / k9s as normal — it's already pointing at the right cluster
k9s
```

## Shell Integration

`awsw` needs to set `AWS_PROFILE` in your **current** shell. Add one of the following to your shell rc file:

**fish** (`~/.config/fish/config.fish`):
```fish
awsw init fish | source
```

**bash** (`~/.bashrc`):
```bash
eval "$(awsw init bash)"
```

**zsh** (`~/.zshrc`):
```zsh
eval "$(awsw init zsh)"
```

## Commands

### `awsw`

Interactive mode. Presents a fuzzy-searchable list via `fzf`. EKS profiles show one line per cluster:

```
$ awsw
  dev-ro                dev (111111111111) ReadOnlyAccess
  my-cluster-dev        dev (111111111111) eng-eks
> my-cluster-staging    staging (222222222222) eng-eks
  prod-ro               production (333333333333) ReadOnlyAccess

→ AWS_PROFILE=it-eks
→ kubectl context: my-cluster-staging
Ready.
```

### `awsw <alias-or-cluster>`

Switch by profile alias **or** cluster name:

```
$ awsw dev-eks
→ AWS_PROFILE=dev-eks
→ kubectl context: my-cluster-dev
Ready.

$ awsw my-cluster-staging
→ AWS_PROFILE=it-eks
→ kubectl context: my-cluster-staging
Ready.
```

### `awsw login`

Authenticate the shared SSO session. Skips the browser if already authenticated. Use `--force` to re-authenticate.

### `awsw status`

Show current state:

```
$ awsw status
AWS Profile:  dev-eks
Account:      111111111111
Role:         eng-eks/nicholas
Kube Context: dev-eks
```

### `awsw setup`

Generate `~/.aws/config` with all SSO profiles and populate `~/.kube/config` with EKS cluster contexts using human-friendly names.

## Profiles

| Alias | Account | ID | Role | EKS |
|---|---|---|---|---|
| `dev-ro` | dev | `111111111111` | ReadOnlyAccess | — |
| `dev-eks` | dev | `111111111111` | eng-eks | ✓ |
| `it-eks` | staging | `222222222222` | eng-eks | ✓ |
| `prod-ro` | production | `333333333333` | ReadOnlyAccess | — |

## Dependencies

- [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [fzf](https://github.com/junegunn/fzf)

## License

This project is licensed under the [PolyForm Shield License 1.0.0](https://polyformproject.org/licenses/shield/1.0.0/) — see [LICENSE](LICENSE) for details.