
### Custom Resources

* [Minecraft](#minecraft)

### Sub Resources

* [MinecraftList](#minecraftlist)
* [MinecraftSpec](#minecraftspec)

#### Minecraft

Minecraft is the Schema for the minecrafts API

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | metav1.ObjectMeta | false |
| spec |  | [MinecraftSpec](#minecraftspec) | false |
| status |  | [MinecraftStatus](#minecraftstatus) | false |

[Back to Custom Resources](#custom-resources)

#### MinecraftList

MinecraftList contains a list of Minecraft

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | metav1.ListMeta | false |
| items |  | [][Minecraft](#minecraft) | true |

[Back to Custom Resources](#custom-resources)

#### MinecraftSpec

MinecraftSpec defines the desired state of Minecraft

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| image | Image for minecraft server | string | false |
| env | Environment variable to be set for the container `EULA` is required The server will not start unless EULA=true. | []corev1.EnvVar | true |
| volumeClaimSpec | PersistentVolumeClaimSpec is a specification of `PersistentVolumeClaim` for persisting data in minecraft. | corev1.PersistentVolumeClaimSpec | true |

[Back to Custom Resources](#custom-resources)
