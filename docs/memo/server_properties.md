```yaml
apiVersion: v1
data:
  difficulty: normal
  motd: A Vanilla Minecraft Server powered by MCing
  pvp: "true"
kind: ConfigMap
metadata:
  name: mcing-server-props
```
↑のようなserver.propertiesをユーザーが用意
Minecraftリソースの.spec.serverPropertiesConfigMapNameにConfigMapの名前を追加しておく

Minecraftリソースのreconcile中(reconcileConfigmap)に
ConfigMapからmap[string]stringとして値を持ってきて、ハードコードしているデフォルト値
package config defaultServerPropsとマージしてserver.propertiesを生成してconfigmapとして保存

mcing-initでconfigmapから作られたファイルを/dataしたにコピー
