# Configuration

## Config File

`awsw` reads its configuration from `~/.config/awsw/config.yaml`. If the file does not exist, built-in defaults are used.

To create the config file with defaults:

```sh
awsw config init
```

Then edit the file to customize your setup.

## Config file format

```yaml
sso:
  start_url: https://d-1234567890.awsapps.com/start/#
  region: us-east-1
  session_name: default
  registration_scopes: sso:account:access
default_region: us-east-1
profiles:
  - alias: dev-ro
    account_name: dev
    account_id: "111111111111"
    role_name: ReadOnlyAccess
  - alias: dev-eks
    account_name: dev
    account_id: "111111111111"
    role_name: eng-eks
    eks_clusters:
      - name: my-cluster-dev
        region: us-east-1
        context_alias: my-cluster-dev
```

## Profile fields

| Field          | Required | Description                                    |
|----------------|----------|------------------------------------------------|
| `alias`        | yes      | Short name used on the CLI (e.g. `dev-eks`)    |
| `account_name` | yes      | Human-readable account name                    |
| `account_id`   | yes      | AWS account number                             |
| `role_name`    | yes      | SSO permission set / role name                 |
| `eks_clusters` | no       | List of EKS clusters associated with this profile |

## EKS cluster fields

| Field           | Required | Description                                    |
|-----------------|----------|------------------------------------------------|
| `name`          | yes      | EKS cluster name                               |
| `region`        | yes      | AWS region where the cluster runs              |
| `context_alias` | yes      | Kubeconfig context alias for `kubectl`         |

## SSO settings

| Field                  | Description                          |
|------------------------|--------------------------------------|
| `sso.start_url`        | AWS SSO start URL                    |
| `sso.region`           | AWS SSO region                       |
| `sso.session_name`     | SSO session name (used with `aws sso login`) |
| `sso.registration_scopes` | SSO registration scopes           |
| `default_region`       | Default AWS region for profiles      |

### Adding or removing a profile

Edit `~/.config/awsw/config.yaml` and add/remove entries under `profiles:`. Changes take effect on the next `awsw` invocation — no rebuild required.

## AWS Config (`~/.aws/config`)

`awsw setup` generates the AWS CLI config file. The generated file contains:

1. A shared `[sso-session default]` block with the SSO start URL and region.
2. One `[profile <alias>]` block per profile, referencing the SSO session.

If `~/.aws/config` already exists, it is backed up to `~/.aws/config.bak` before being overwritten.

### SSO session settings

Configured in the `sso:` section of the config file. Defaults:

- **Start URL**: `https://d-1234567890.awsapps.com/start/#`
- **SSO Region**: `us-east-1`
- **Registration scopes**: `sso:account:access`

## Kubeconfig (`~/.kube/config`)

`awsw setup` configures kubeconfig for every profile that has `eks_clusters` defined. For each cluster it runs:

```sh
aws eks update-kubeconfig --name <cluster-name> --region <region> --profile <alias> --alias <context-alias>
```

Cluster names, regions, and context aliases are all explicitly configured — no auto-discovery is performed.

## Environment Variables

`awsw` reads and sets the following environment variables:

| Variable       | Read/Set | Description                        |
|----------------|----------|------------------------------------|
| `AWS_PROFILE`  | Set      | Active AWS CLI profile name        |
| `SHELL`        | Read     | Fallback for detecting shell type  |

## External Dependencies

These must be installed and on `$PATH`:

- **AWS CLI v2** — SSO login, STS identity, EKS operations
- **kubectl** — context switching, kubeconfig management
- **fzf** — interactive fuzzy selection (only needed for interactive mode)
