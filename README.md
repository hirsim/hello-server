# Hello Server
ただ`Hello World!`を返すWebサーバです。

## エンドポイント

エンドポイント | メソッド | レスポンス
--- | --- | ---
/ | GET | `{"message":"Hello World!"}`
/hello | GET | `Hello World!`
/hello.html | GET | `<h1>Hello World!</h1>` *HTML5*
/hello.json | GET | `{"message":"Hello World!"}`
/sloth/hello | GET | `Hello World!` (30秒レスポンスを保留します)
/sloth/hello.html | GET | `<h1>Hello World!</h1>` *HTML5* (30秒レスポンスを保留します)
/sloth/hello.json | GET | `{"message":"Hello World!"}` (30秒レスポンスを保留します)

## コンテナビルド
multi-stage builds を使用しているため、 17.05 以降のバージョンを使用してください。

```
docker build -t hello-server:latest .
```

## コンテナ起動

```
docker run -d -p 8080:8080 hello-server:latest
curl http://localhost:8080/hello
Hello World!
```

環境変数 PRINT_TEXT に文字列を設定すると、その文字列を表示します。

```
docker run -d -e PRINT_TEXT="HELLO WORLD!" -p 8080:8080 hello-server:latest
curl http://localhost:8080/hello
HELLO WORLD!
```

## ライセンス

[CC0-1.0](./LICENSE)