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

    MCing does not yet have the ability to customize the Service, so you can use port-forward and other methods to check access to the server.
    ```console
    $ kubectl port-forward svc/minecraft-sample 25565:25565
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
However, these files are updated from the command in actual operation. Therefore, they will be applied if the target files are not present at startup.
(TODO: I'm trying to figure out how to successfully merge the JSON specified in ConfigMap with the JSON that actually exists).
