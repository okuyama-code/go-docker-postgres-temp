## git cloneしたときにやること
```
docker compose up --build
```

ビルド時にキャッシュを使用したくない場合
```
docker compose build --no-cache
```

## 起動 app
```
docker compose up
```

サンプルプログラムで動作の確認
app コンテナに入る
```
docker compose exec app bash
```
```
go run main.go
```
postmanなどを使ってこれらを叩いてみる
http://localhost:8080/register
http://localhost:8080/login


メソッド: POST
URL: http://localhost:8080/register
http://localhost:8080/login こっちも同じように。


ヘッダー設定:

Key: Content-Type
Value: application/json


ボディ設定:

「raw」を選択し、JSONフォーマットを指定
以下のJSONを入力:
```
{
  "username": "testuser",
  "password": "testpassword123"
}
```

送信ボタンをクリックしてリクエストを送信


期待される応答:
/register

```
{
    "ID": 2,
    "CreatedAt": "2024-06-23T09:46:28.614542586Z",
    "UpdatedAt": "2024-06-23T09:46:28.614542586Z",
    "DeletedAt": null,
    "username": "okuyama22",
    "password": "$2a$10$g0TVCV/A3Jfv2dKtVaQVXuA5y/TDC.SIeYmoLJSSDi16SeXY3TX/W"
}
```

/loginの成功は
```
成功の場合 (HTTPステータスコード 200):
jsonCopy{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```



## postgresの中に入ってテーブルなどを確認するために
```
docker compose up
```

db コンテナに入る
```
docker compose exec db bash
```
コンテナ内のシェルに入ったら、PostgreSQLのコマンドラインツールであるpsqlを使用してデータベースに接続します。
psql -U <ユーザー名> -d <データベース名>

```
psql -U postgres -d react_go_app
```

テーブル一覧を表示 (cloneしたてはなにもない)
```
\dt
```
特定のテーブルを確認　\d <テーブル名>　カラムなどを見れる

```
\d users
```
テーブルのデータを確認
```
SELECT * FROM users;
```

PostgreSQLのプロンプトを終了する \はoption + ¥
```
\q
```

## DB migrate drop migrate:reset
コンテナに入ってから
```
docker compose exec app bash
```

migrate
```
go run main.go -migrate
```

drop
```
go run main.go -drop
```

migrate:reset
```
go run main.go -reset
```


