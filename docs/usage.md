# Usage

1. Apply Minecraft Custom Resouce

    ```console
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
    $ kubectl apply -f minecraft-sample.yaml

    # check resources 
    $ kubectl get minecrafts.mcing.kmdkuk.com minecraft-sample
    NAME               AGE
    minecraft-sample   2m41s

    $ kubectl get statefulsets.apps minecraft-sample
    NAME               READY   AGE
    minecraft-sample   1/1     2m47s

    $ kubectl get persistentvolumeclaims minecraft-data-minecraft-sample-0
    NAME                                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    minecraft-data-minecraft-sample-0   Bound    pvc-220c9ba1-2395-48ed-a18a-c9d3d7788a6c   1Gi        RWO            standard       3h58m

    $ kubectl get service minecraft-sample
    NAME               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)               AGE
    minecraft-sample   ClusterIP   10.96.213.69   <none>        25565/TCP,25575/UDP   2m58s
    ```

    When you apply Minecraft Custom Resource, several resources will be created.

2. Access the server

    MCing does not yet have the ability to customize the Service, so you can use port-forward and other methods to check access to the server.
    ```console
    $ kubectl port-forward svc/minecraft-sample 25565:25565
    ```
