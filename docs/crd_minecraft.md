
### Custom Resources

* [Minecraft](#minecraft)

### Sub Resources

* [MinecraftList](#minecraftlist)
* [MinecraftSpec](#minecraftspec)
* [ObjectMeta](#objectmeta)
* [PodSpec](#podspec)
* [PodTemplateSpec](#podtemplatespec)

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
| volumeClaimTemplates | VolumeClaimTemplates is a list of `PersistentVolumeClaim` templates for Minecraft server container. A claim named \"minecraft-data\" must be included in the list. | []corev1.PersistentVolumeClaim | true |

[Back to Custom Resources](#custom-resources)

#### ObjectMeta

ObjectMeta is metadata of objects. This is partially copied from metav1.ObjectMeta.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name is the name of the object. | string | false |
| labels | Labels is a map of string keys and values. | map[string]string | false |
| annotations | Annotations is a map of string keys and values. | map[string]string | false |

[Back to Custom Resources](#custom-resources)

#### PodSpec

PodSpec is a description of a pod. This is slightly modified from corev1.PodSpec.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object's metadata.  The name in this metadata is ignored. | [ObjectMeta](#objectmeta) | false |
| containers | List of containers belonging to the pod. The name of the Minecraft server container in this Containers must be `minecraft`. | []corev1.Container | true |
| volumes | List of volumes that can be mounted by containers belonging to the pod. | []corev1.Volume | true |
| serviceSpec | Specification of the desired behavior of the service. | corev1.ServiceSpec | false |
| minecraftConfigMapName | MinecraftConfigMapName is a `ConfigMap` name of Minecraft server config. | *string | false |

[Back to Custom Resources](#custom-resources)

#### PodTemplateSpec

PodTemplateSpec describes the data a pod should have when created from a template. This is slightly modified from corev1.PodTemplateSpec.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object's metadata.  The name in this metadata is ignored. | [ObjectMeta](#objectmeta) | false |
| spec | Specification of the desired behavior of the pod. | [PodSpec](#podspec) | true |

[Back to Custom Resources](#custom-resources)
