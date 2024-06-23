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
docker-compose exec app sh
```
```
go run main.go
```
Testingと出たら動作している

## postgresの中に入ってテーブルなどを確認するために
```
docker compose up
```

db コンテナに入る
```
docker-compose exec db bash
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
特定のテーブルを確認
```
\d <テーブル名>
```

PostgreSQLのプロンプトを終了する \はoption + ¥
```
\q
```


