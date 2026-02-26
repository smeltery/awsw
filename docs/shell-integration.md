# Shell Integration

## Why a shell wrapper?

`awsw` needs to set `AWS_PROFILE` in your current shell session. A normal binary runs in a subprocess and cannot modify its parent's environment. The solution is a shell function that `eval`s the binary's output.

This is the same pattern used by `direnv`, `mise`, `starship`, and `rbenv`.

## Setup

### Fish

Add to `~/.config/fish/config.fish`:

```fish
awsw init fish | source
```

### Bash

Add to `~/.bashrc`:

```bash
eval "$(awsw init bash)"
```

### Zsh

Add to `~/.zshrc`:

```zsh
eval "$(awsw init zsh)"
```

## How it works

`awsw init <shell>` prints a shell function to stdout. For example, `awsw init fish` outputs:

```fish
function awsw
    set -l cmd (command /path/to/awsw --shell-eval --shell fish $argv)
    if test $status -eq 0
        eval $cmd
    end
end
```

When you run `awsw dev-eks`, this function:
1. Calls the real binary with `--shell-eval --shell fish dev-eks`
2. The binary resolves the alias, then prints `set -gx AWS_PROFILE dev-eks`
3. The function `eval`s that output, setting `AWS_PROFILE` in your shell

Commands that don't change env vars (`login`, `status`, `setup`) are passed through directly — the binary handles them without eval.

## Verifying

After adding the init hook, restart your shell and run:

```sh
type awsw
```

You should see a function definition, not a binary path. If you see the binary path, the init hook isn't being sourced.

## Troubleshooting

**`awsw: command not found` after init**
The binary isn't on your `$PATH`. Install it first, then add the init hook.

**`AWS_PROFILE` not changing**
Make sure you're using the shell wrapper (function), not calling the binary directly. Check with `type awsw`.

**Wrong shell detected**
The init command takes an explicit shell argument. Make sure it matches your actual shell (e.g. don't put `awsw init bash` in your `.zshrc`).
