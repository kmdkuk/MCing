# Deploying MCing

1. (Optional) Prepare cert-manager

    MCing depends on [cert-manager][] to issue TLS certificate for admission webhooks.
    If cert-manager is not installed on your cluster, install it as follows:

    ```console
    $ curl -fsLO https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
    $ kubectl apply -f cert-manager.yaml
    ```

2. Apply MCing manifests

    Please Install [kustomize/v4.1.3](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv4.1.3)
    ```console
    $ curl -fsLO https://github.com/kmdkuk/MCing/releases/latest/download/install.yaml
    $ kubectl apply -f install.yaml
    ```

[cert-manager]: https://cert-manager.io/
