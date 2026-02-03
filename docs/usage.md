# Usage

## Starting simple server

1. Apply Minecraft Custom Resouce

    ```console
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
    $ kubectl apply -f minecraft-sample.yaml

    # check resources 
    $ kubectl get minecrafts.mcing.kmdkuk.com minecraft-sample
    NAME               AGE
    minecraft-sample   2m41s

    $ kubectl get statefulsets.apps mcing-minecraft-sample
    NAME               READY   AGE
    mcing-minecraft-sample   1/1     2m47s

    $ kubectl get persistentvolumeclaims minecraft-data-minecraft-sample-0
    NAME                                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    minecraft-data-mcing-minecraft-sample-0   Bound    pvc-220c9ba1-2395-48ed-a18a-c9d3d7788a6c   1Gi        RWO            standard       3h58m

    $ kubectl get service
    NAME                              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                           AGE
    kubernetes                        ClusterIP   10.96.0.1      <none>        443/TCP                           12m
    mcing-minecraft-sample            NodePort    10.96.232.50   <none>        25565:30088/TCP,25575:30446/UDP   103s
    mcing-minecraft-sample-headless   ClusterIP   None           <none>        25565/TCP,25575/UDP               103s

    $ kubectl get configmaps
    NAME                     DATA   AGE
    kube-root-ca.crt         1      13m
    mcing-minecraft-sample   1      2m47s
    mcing-server-props       3      2m47s
    ```

    When you apply Minecraft Custom Resource, several resources will be created.

2. Access the server

    You can customize the Service (e.g. change type to LoadBalancer) using `ServiceTemplate` in the Minecraft Custom Resource to access the server.
    Alternatively, you can use port-forward to verify access quickly.

    ```console
    kubectl port-forward svc/minecraft-sample 25565:25565
    ```

## Configuration by ConfigMap

If you edit the ConfigMap specified by `.spec.serverPropertiesConfigMapName` in the Minecraft resource, it will automatically replace server.properties and then execute the `/reload` command.
Note: There are some cases where the configuration will not be updated even if you run the `/reload` command. (In the case of TYPE=SPIGOT, we have confirmed that the configuration is updated automatically.)

The following ConfigMap can be applied to other JSON configuration files.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: other-props
data:
  banned-ips.json: |
    []
  banned-players.json: |
    []
  whitelist.json: |
    []
  ops.json: |
    []
```

It can be applied by specifying a name, such as `otherConfigMapName: other-props`.

> [!WARNING]
> Currently, these files are **overwritten** by the ConfigMap content on every Pod startup.
> This means any in-game changes (e.g. `/ban`, `/op`) that are not reflected in the ConfigMap will be lost when the Pod restarts.
> This behavior is a known issue.

## RCON Configuration

By default, the controller automatically generates a Secret named `<instance-name>-rcon-password` containing a random password for RCON.
The password is stored in the key `rcon-password`.
The controller injects this password into the Minecraft container as the environment variable `RCON_PASSWORD`.

If you want to use your own password, you can specify an existing Secret name in `.spec.rconPasswordSecretName`.

```yaml
apiVersion: mcing.kmdkuk.com/v1alpha1
kind: Minecraft
metadata:
  name: minecraft-sample
spec:
  rconPasswordSecretName: my-rcon-secret
  # ... other fields
```

The Secret must contain the key `rcon-password`.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-rcon-secret
stringData:
  rcon-password: "your-super-strong-password"
```

## Auto-Pause

MCing supports automatic server pausing when no players are connected, using [lazymc](https://github.com/timvisee/lazymc). This helps reduce resource usage for idle servers.

### Enabling Auto-Pause

```yaml
apiVersion: mcing.kmdkuk.com/v1alpha1
kind: Minecraft
metadata:
  name: minecraft-sample
spec:
  autoPause:
    enabled: true
    timeoutSeconds: 300  # Pause after 5 minutes of inactivity (default)
  # ... other fields
```

### How it works

1. When enabled, lazymc runs as the main process and proxies connections to the Minecraft server
2. The Minecraft server starts when a player connects
3. After `timeoutSeconds` of no player activity, the server is paused
4. The server automatically resumes when a new player connects

> [!NOTE]
> Auto-pause is enabled by default. Set `autoPause.enabled: false` to disable it.

## Operators and Whitelist

MCing can manage operators and whitelist through the Minecraft CR spec.

### Operators

```yaml
spec:
  ops:
    users:
      - player1
      - player2
```

The controller executes `/op` or `/deop` commands via RCON to sync the operators list.

### Whitelist

```yaml
spec:
  whitelist:
    enabled: true
    users:
      - allowed_player1
      - allowed_player2
```

When `whitelist.enabled` is `true`, the controller executes `/whitelist on` and manages the whitelist via `/whitelist add` and `/whitelist remove` commands.

## Backup and Download

MCing provides a kubectl plugin for downloading server data.

### Installing kubectl-mcing

```console
go install github.com/kmdkuk/mcing/cmd/kubectl-mcing@latest
```

Or download from [GitHub Releases](https://github.com/kmdkuk/mcing/releases).

### Downloading Server Data

```console
kubectl mcing download <minecraft-name> [-o output.tar.gz] [-n namespace]
```

This command:

1. Executes `save-off` to disable auto-save
2. Executes `save-all flush` to ensure all data is written
3. Compresses and downloads the `/data` directory
4. Executes `save-on` to re-enable auto-save

### Excluding Files from Backup

You can exclude files from the backup using the `backup.excludes` field:

```yaml
spec:
  backup:
    excludes:
      - "*.jar"
      - "logs/*"
      - "cache/*"
```

## mc-router (Hostname-based Routing)

When mc-router is enabled on the controller, you can use custom hostnames to access your Minecraft servers.

### Using Custom Hostname

```yaml
spec:
  externalHostname: "survival.mc.example.com"
```

If not specified, the hostname is automatically generated as `<name>.<namespace>.<default-domain>`.

### DNS Configuration

Point a wildcard DNS record to the mc-router service's external IP:

```text
*.mc.example.com -> <mc-router-external-ip>
```

Players can then connect using hostnames like `survival.mc.example.com:25565`.

See [Deploy MCing](setup.md#enabling-mc-router-hostname-based-routing) for controller configuration.
