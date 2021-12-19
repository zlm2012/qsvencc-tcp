# TCP Wrapper for QSVEncC

このプロジェクトは、[QSVEnc](https://github.com/rigaya/QSVEnc) にある QSVEncC の TCP Wrapper です。
単純にクライアントから入力とパラメタをサーバに送って、サーバから出力と進捗など Stderr 出力をクライアントに返す実装です。

(C) 2021 Lingmo Zhu

## 考えられる使い道

* QSV でエンコがほしいですが QSV 対応パソコンは別です (あるかな？)
* VM に QSV を使いたいですが、GVT-g か適切な SR-IOV が対応しいてないから VM から QSV を使うのが難しいか無理 (例えば11世代内蔵 Xe Graphics)

## コンポネート

* `cmd/qsvencc-server` サーバの実装で、下記のオプションが対応しています
  * `-l --listen`: listen 用 tcp アドレス、デフォルトは `192.168.122.1:11111`
  * `-c --cmd-path`: qsvencc までのファイルパス、デフォルトは `qsvencc`
* `cmd/qsvencc-client` クライアントの実装で、qsvencc のパラメタそのままでもほぼ使えます。下記のオプションが対応しています
  * `--remote-addr`: 接続用 tcp アドレス, デフォルトは `192.168.122.1:11111`

なぜ`192.168.122.1`をデフォルトに? KVM のデフォルト net bridge `virbr0` にデフォルトのホストアドレスはこれですから。

## KVM との利用の注意点

一部設定を調整しないと動作が重くて許容できないかも：

* raw 動画のファイルかストリームが入力の場合 virtio NIC は強く推奨します
* ゲストの CPU 設定に少なくとも `mode='host-model'` までしてください、でないと CPU の性能が大幅に下がる恐れがあります

## License

MIT License は適用しています。詳細は [LICENSE](./LICENSE) で確認ください。