# Deploying MCing

## Basic Installation

1. (Optional) Prepare cert-manager

    MCing depends on [cert-manager][] to issue TLS certificate for admission webhooks.
    If cert-manager is not installed on your cluster, install it as follows:

    ```console
    $ curl -fsLO https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
    $ kubectl apply -f cert-manager.yaml
    ```

2. Apply MCing manifests

    ```console
    $ curl -fsLO https://github.com/kmdkuk/MCing/releases/latest/download/install.yaml
    $ kubectl apply -f install.yaml
    ```

## Customizing Controller Settings

To customize the MCing controller, download and edit `install.yaml` before applying:

```console
$ curl -fsLO https://github.com/kmdkuk/MCing/releases/latest/download/install.yaml
# Edit install.yaml to add/modify controller flags
$ kubectl apply -f install.yaml
```

### Controller Flags

The MCing controller supports the following flags:

| Flag                           | Default  | Description                                     |
| ------------------------------ | -------- | ----------------------------------------------- |
| `--metrics-bind-address`       | `:8080`  | The address the metric endpoint binds to        |
| `--health-probe-bind-address`  | `:8081`  | The address the probe endpoint binds to         |
| `--leader-elect`               | `false`  | Enable leader election for HA                   |
| `--check-interval`             | `1m`     | Interval of Minecraft server maintenance checks |

### Enabling mc-router (Hostname-based Routing)

mc-router allows multiple Minecraft servers to share a single external IP/port by routing based on hostname.

Add the following flags to the controller deployment:

| Flag                         | Default                 | Description                                                        |
| ---------------------------- | ----------------------- | ------------------------------------------------------------------ |
| `--enable-mc-router`         | `false`                 | Enable mc-router gateway                                           |
| `--mc-router-default-domain` | `minecraft.local`       | Default domain for FQDN generation (`<name>.<namespace>.<domain>`) |
| `--mc-router-namespace`      | `mcing-gateway`         | Namespace for mc-router deployment                                 |
| `--mc-router-service-type`   | `LoadBalancer`          | Service type (`LoadBalancer` or `NodePort`)                        |
| `--mc-router-image`          | `itzg/mc-router:latest` | mc-router container image                                          |

Example deployment patch:

```yaml
spec:
  template:
    spec:
      containers:
      - name: manager
        args:
        - --leader-elect
        - --enable-mc-router
        - --mc-router-default-domain=mc.example.com
        - --mc-router-service-type=LoadBalancer
```

When mc-router is enabled:

- Each Minecraft server gets a hostname: `<name>.<namespace>.<default-domain>` (or custom `externalHostname`)
- Services are created as ClusterIP type and routed through mc-router
- Point your DNS wildcard (`*.mc.example.com`) to the mc-router service

[cert-manager]: https://cert-manager.io/
