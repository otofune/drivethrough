drivethrough
===

ドライブにあるファイルを HTTP で直接参照できるようにする

つかいかた
---

1. Google Drive API を [Google Developer Console](console.developers.google.com/) で有効にする
1. OAuth2 アプリをその他 (Others) カテゴリで作成する
1. 実行する
    ```sh
    git clone https://github.com/otofune/drivethrough
    cd drivethrough
    export DT_GOOGLE_CLIENT_ID='your client ID'
    export DT_GOOGLE_CLIENT_SECRET='your client Secret'
    go run .
    ```
1. curl "http://localhost:10000/example.txt" でドライブ直下にある "example.txt" が取れるようになる
1. ENJOY

TODO
---
- [x] `application/vnd.google-apps.shortcut` に対応する
- [ ] ディレクトリのキャッシュに期限を持つようにする
- [ ] zap 入れてデバッグログを出せるようにする
- [ ] 返却する値に Content-Length, Content-Type を付ける
- [ ] Range リクエストをサポートする
- [ ] エラーメッセージや README を英語にする
- [ ] ディレクトリの際にファイルリストを返却する
- [ ] 同じ名前を持つディレクトリファイルが複数あった場合すぐにエラーにせず子ディレクトリまで検知してみる
