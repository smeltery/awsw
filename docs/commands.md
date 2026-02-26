# Command Reference

## `awsw` (default — switch)

Switch AWS profile and kubectl context.

### Interactive mode

```sh
awsw
```

Opens an `fzf` selector. EKS profiles show one line per cluster (keyed by context alias); non-EKS profiles show one line per profile. On selection, sets `AWS_PROFILE` and switches the kubectl context.

### Direct switch

```sh
awsw <alias-or-cluster>
```

Switches directly by **profile alias** or **EKS cluster name**. Resolution order:

1. Exact match on profile alias
2. Exact match on EKS cluster context alias
3. Exact match on EKS cluster name

Examples:

```sh
awsw dev-eks              # → AWS_PROFILE=dev-eks, context=my-cluster-dev
awsw my-cluster-staging   # → AWS_PROFILE=it-eks, context=my-cluster-staging
awsw prod-ro              # → AWS_PROFILE=prod-ro (no context switch, no EKS)
```

### Available aliases

- `dev-ro` — dev account, ReadOnlyAccess
- `dev-eks` — dev account, eng-eks role + EKS context
- `it-eks` — staging account, eng-eks role + EKS context
- `prod-ro` — production account, ReadOnlyAccess

You can also use any configured EKS cluster name or context alias directly.

---

## `awsw login`

Authenticate the shared SSO session.

```sh
awsw login
awsw login --force   # re-authenticate even if session is valid
```

Before opening the browser, checks if the current session is already valid via `aws sts get-caller-identity`. If authenticated, prints a message and exits. Use `--force` to re-authenticate regardless.

A single login grants access to all accounts and roles.

---

## `awsw status`

Show the current state.

```sh
awsw status
```

Prints:
- Current `AWS_PROFILE` value
- `aws sts get-caller-identity` output (account, ARN, user ID)
- Current kubectl context (`kubectl config current-context`)

---

## `awsw setup`

First-time setup or refresh.

```sh
awsw setup
```

Performs two operations:

1. **AWS config generation** — writes `~/.aws/config` with SSO session and profile blocks. Backs up any existing file to `~/.aws/config.bak`.
2. **EKS kubeconfig population** — for each EKS-enabled profile, discovers clusters via `aws eks list-clusters` and registers them in `~/.kube/config` with friendly context aliases.

Run this once after installation, or again if the account registry changes.

---

## `awsw init <shell>`

Emit the shell wrapper function.

```sh
awsw init fish   # for fish shell
awsw init bash   # for bash
awsw init zsh    # for zsh
```

Prints a shell function to stdout. Add it to your shell rc file — see [shell-integration.md](shell-integration.md) for setup instructions.

---

## `awsw config init`

Scaffold a config file with built-in defaults.

```sh
awsw config init
awsw config init --force   # overwrite existing
```

Writes the default configuration to `~/.config/awsw/config.yaml`. Refuses to overwrite an existing file unless `--force` is passed.

See [configuration.md](configuration.md) for the config file format.

---

## Global Flags

| Flag           | Description                                      |
|----------------|--------------------------------------------------|
| `--shell-eval` | Output shell commands instead of executing them (used internally by the shell wrapper) |
| `--shell`      | Specify the shell type (`fish`, `bash`, `zsh`)   |
| `--version`    | Print the version and exit                       |
