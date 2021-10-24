[![CI](https://github.com/kmdkuk/mcing/actions/workflows/ci.yaml/badge.svg)](https://github.com/kmdkuk/mcing/actions/workflows/ci.yaml)

# MCing


MCing is a Kubernetes operator for Minecraft server.

## Supported software

- Support Server Image: [itzg/minecraft-server](https://hub.docker.com/r/itzg/minecraft-server)
- Kubernetes: 1.19, 1.20, 1.21

## Quick start

You can quickly run MCing using [kind](https://kind.sigs.k8s.io/).

```
$ cd e2e
$ make start bootstrap
$ cat <<EOF > minecraft-sample.yaml
apiVersion: mcing.kmdkuk.com/v1alpha1
kind: Minecraft
metadata:
    name: minecraft-sample
spec:
    env:
        - name: EULA
        value: "true"
    volumeClaimSpec:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: standard
        resources:
        requests:
            storage: 1Gi
EOF
$ ../bin/kubectl --kubeconfig .kubeconfig apply -f minecraft-sample.yaml
$ ../bin/kubectl --kubeconfig .kubeconfig port-forward svc/minecraft-sample 25565:25565
```

## Documentation

See https://kmdk.uk/MCing

## Docker images

Docker images are available on [ghcr.io/kmdkuk/mcing](https://github.com/kmdkuk/packages/container/package/mcing).
