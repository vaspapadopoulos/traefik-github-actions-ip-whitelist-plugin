# Traefik Github Actions IP Whitelist Plugin

[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://github.com/pre-commit/pre-commit)

A middleware plugin for Traefik that whitelists TCP connections from Github
Actions or other specified addresses. The Github Actions addresses are
automatically retrieved from the
[Gitbub Meta information endpoint](https://docs.github.com/en/rest/reference/meta#get-github-meta-information)
.

## Configuration

```yaml
testData:
  additionalCIDRs:
    - 13.67.144.0/21
    - 13.67.152.0/24
    - 13.67.153.0/28
```

- `additionalCIDRs` additional CIDRs to be added in the whitelist

### Example configuration

- Static configuration

```yaml
pilot:
  token: <token>

experimental:
  plugins:
    github-actions-ip-whitelist:
      moduleName: github.com/vaspapadopoulos/traefik-github-actions-ip-whitelist-plugin
      version: v0.1.0
```

- Dynamic configuration

```yaml
tcp:
  routers:
    my-service.com:
      service: my-service
      middlewares:
        - githubActionsIpWhitelist
  middlewares:
    githubActionsIpWhitelist:
      plugin:
        github-actions-ip-whitelist:
          additionalCIDRs:
            - 13.67.144.0/21
            - 13.67.152.0/24
            - 13.67.153.0/28
  services:
    my-service:
      loadBalancer:
        servers:
          - url: <url>
```
