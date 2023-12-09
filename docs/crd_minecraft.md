
### Custom Resources

* [Minecraft](#minecraft)

### Sub Resources

* [MinecraftList](#minecraftlist)
* [MinecraftSpec](#minecraftspec)
* [ObjectMeta](#objectmeta)
* [Ops](#ops)
* [PersistentVolumeClaim](#persistentvolumeclaim)
* [PodTemplateSpec](#podtemplatespec)
* [ServiceTemplate](#servicetemplate)
* [Whitelist](#whitelist)

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
| podTemplate | PodTemplate is a `Pod` template for Minecraft server container. | [PodTemplateSpec](#podtemplatespec) | true |
| volumeClaimTemplates | PersistentVolumeClaimSpec is a specification of `PersistentVolumeClaim` for persisting data in minecraft. A claim named \"minecraft-data\" must be included in the list. | [][PersistentVolumeClaim](#persistentvolumeclaim) | true |
| serviceTemplate | ServiceTemplate is a `Service` template. | *[ServiceTemplate](#servicetemplate) | false |
| ops | operators on server. exec /op or /deop | [Ops](#ops) | false |
| whitelist | whitelist | [Whitelist](#whitelist) | false |
| serverPropertiesConfigMapName | ServerPropertiesConfigMapName is a `ConfigMap` name of `server.properties`. | *string | false |
| otherConfigMapName | OtherConfigMapName is a `ConfigMap` name of other configurations file(eg. banned-ips.json, ops.json etc) | *string | false |

[Back to Custom Resources](#custom-resources)

#### ObjectMeta

ObjectMeta is metadata of objects. This is partially copied from metav1.ObjectMeta.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name is the name of the object. | string | false |
| labels | Labels is a map of string keys and values. | map[string]string | false |
| annotations | Annotations is a map of string keys and values. | map[string]string | false |

[Back to Custom Resources](#custom-resources)

#### Ops



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| users | user name exec /op or /deop | []string | false |

[Back to Custom Resources](#custom-resources)

#### PersistentVolumeClaim

PersistentVolumeClaim is a user's request for and claim to a persistent volume. This is slightly modified from corev1.PersistentVolumeClaim.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object's metadata. | [ObjectMeta](#objectmeta) | true |
| spec | Spec defines the desired characteristics of a volume requested by a pod author. | corev1.PersistentVolumeClaimSpec | true |

[Back to Custom Resources](#custom-resources)

#### PodTemplateSpec

PodTemplateSpec describes the data a pod should have when created from a template. This is slightly modified from corev1.PodTemplateSpec.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object's metadata.  The name in this metadata is ignored. | [ObjectMeta](#objectmeta) | false |
| spec | Specification of the desired behavior of the pod. The name of the MySQL server container in this spec must be `minecraft`. | corev1.PodSpec | true |

[Back to Custom Resources](#custom-resources)

#### ServiceTemplate

ServiceTemplate define the desired spec and annotations of Service

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object's metadata. Only `annotations` and `labels` are valid. | [ObjectMeta](#objectmeta) | false |
| spec | Spec is the ServiceSpec | *corev1.ServiceSpec | false |

[Back to Custom Resources](#custom-resources)

#### Whitelist



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enabled | exec /whitelist on | bool | false |
| users | user name exec /whitelist add or /whitelist remove | []string | false |

[Back to Custom Resources](#custom-resources)
