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
    podTemplate:
    spec:
        containers:
        - name: minecraft
            image: itzg/minecraft-server:java8
            env:
            - name: TYPE
                value: "SPIGOT"
            - name: EULA
                value: "true"
    serviceTemplate:
    spec:
        type: NodePort
    volumeClaimTemplates:
    - metadata:
        name: minecraft-data
        spec:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: standard
        resources:
            requests:
            storage: 1Gi
    serverPropertiesConfigMapName: mcing-server-props
---
apiVersion: v1
kind: ConfigMap
metadata:
    name: mcing-server-props
data:
    motd: "[this is test]A Vanilla Minecraft Server powered by MCing"
    pvp: "true"
    difficulty: "hard"
EOF
EOF
$ ../bin/kubectl --kubeconfig .kubeconfig apply -f minecraft-sample.yaml
$ ../bin/kubectl --kubeconfig .kubeconfig port-forward svc/minecraft-sample 25565:25565
```

## Documentation

See https://kmdk.uk/MCing

## Docker images

Docker images are available on [ghcr.io/kmdkuk/mcing](https://github.com/kmdkuk/packages/container/package/mcing).
