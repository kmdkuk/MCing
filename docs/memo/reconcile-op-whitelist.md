Minecraftリソースに.spec.whitelist .spec.opsを追加

reconcile時にwhitelistの追加, opsの追加を行う。
ただ、初回起動時にはうまくreconcileの中では追加できないのでreconcileがエラーループしてしまう。
そのためMinecraftリソースごとにgoroutineを立てて、その中でwhitelistとopsの同期をするような処理を考えたい。

reconciler
1. MinecraftリソースApply, reconciler起動
2. goroutineを立てて、何かしら管理処理を定期実行する

ref: https://minecraft.fandom.com/ja/wiki/%E3%82%B3%E3%83%9E%E3%83%B3%E3%83%89/whitelist
goroutine(whitelist)
1. .spec.whitelistをmcing-agentに送信
1. whitelist.enabledがtrueなら, /whitelist on相当の操作、現状のonのチェックだけができるならチェックのみ
1. mcing-agentで受け取ったUsersと/whitelist listと比較
    1. Usersに存在するが、/whitelist listに存在しないものは、/whitelist add 相当の操作をする
    2. /whitelistに存在するが、Usersに存在しないものは、 /whitelist remove 相当の操作をする
1. whitelist add|remove を実行していれば whitelist reloadを実行して終了


opとdeopをwhitelistみたいによしなにsyncしようとしたが、どうもoperator権限を持っている人をリストするようなことが出来ないっぽい。
やるならops.jsonをparseする？
goroutine(ops)
1. .spec.opsをmcing-agentに送信
1. /data/ops.jsonをパース(/dataはすでにmcing-agentにもマウントしている)
1. ops.jsonに存在しないが、.spec.opsに存在するユーザーを /op
1. ops.jsonに存在するが、.spec.opsに存在しないユーザーを /deop
1. 終了 /op /deopにはリロードの概念がなさそう。

必要なもの
- reconcileに、goroutineを起動・停止する処理
- goroutineに、mcing-agentに対してwhitelistとopのリクエストを送る
- mcing-agentに、whitelistとopのリクエストを受け取るprot適宜
  - 処理も記述
- e2eテストとかで、うまいこと編集して権限追加されているか確認するようなテストがあると良さそう。
　　- もっと小さいサイズでテストしたいが、mockとか面倒か？
