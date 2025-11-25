# healthplanet-to-fitbit

もりはや改良版

[![Test](https://github.com/morihaya/healthplanet-to-fitbit/actions/workflows/test.yml/badge.svg)](https://github.com/morihaya/healthplanet-to-fitbit/actions/workflows/test.yml)

[HealthPlanet](https://www.healthplanet.jp/)に登録された体重・体脂肪率情報を Fitbit へ転送する。

## 環境変数

初期設定ツール（`healthplanet-gettoken`, `fitbit-gettoken`）を実行するために、`.env`ファイルまたは環境変数に以下を定義します。
これらのツールを実行すると、アクセストークンなどが取得され、設定ファイル（`~/.config/healthplanet-to-fitbit/config.json`）に保存されます。
メインの同期ツール（`healthplanet-to-fitbit`）は、この `config.json` を使用して動作します。

| 環境変数名                 | 内容                                                        |
| -------------------------- | ----------------------------------------------------------- |
| HEALTHPLANET_CLIENT_ID     | HealthPlanet 公式サイトで発行できるクライアント ID          |
| HEALTHPLANET_CLIENT_SECRET | HealthPlanet 公式サイトで発行できるクライアントシークレット |
| FITBIT_CLIENT_ID           | Fitbit 公式サイトで発行できるクライアント ID                |
| FITBIT_CLIENT_SECRET       | Fitbit 公式サイトで発行できるクライアントシークレット       |

## 事前準備

1. HealthPlant, Fitbit の公式サイトから各種 API キーを取得し、`.env` ファイル等で環境変数に登録する。
2. 以下のコマンドを実行し、HealthPlanet のトークンを取得・保存する。
   ```bash
   go run cmd/healthplanet-gettoken/main.go
   ```
3. 以下のコマンドを実行し、Fitbit のトークンを取得・保存する。
   ```bash
   go run cmd/fitbit-gettoken/main.go
   ```

上記を実行すると、`~/.config/healthplanet-to-fitbit/config.json` に認証情報が保存されます。

## 使用方法

`healthplanet-to-fitbit` を実行する。

```bash
go run cmd/healthplanet-to-fitbit/main.go
```

期間を指定して同期する場合:
```bash
go run cmd/healthplanet-to-fitbit/main.go --from 2025-01-01 --to 2025-01-31
```
処理済みのレコードは `~/.config/healthplanet-to-fitbit/cache.json` にキャッシュされ、次回以降はスキップされます。

設定ファイルから認証情報を読み込み、直近３か月の情報（体重・体脂肪率）が HeathPlanet から取得され、Fitbit へ登録される。
Fitbit のアクセストークンが期限切れの場合は、自動的にリフレッシュされ、設定ファイルが更新される。

## API制限について

HealthPlanet API には **60回/時** 程度の厳しいレートリミットがあるようです（[公式ドキュメント](https://www.healthplanet.jp/apis/api.html)には明記されていませんが、短時間に多数のリクエストを送ると `400 Bad Request (Error 401)` が返ることがあります）。

大量のデータを同期しようとしてエラーが発生した場合は、1時間ほど待ってから再度実行してください。

## テスト

以下のコマンドで単体テストを実行できます。

```bash
go test ./...
```

## References

- [HealthPlanet API](https://www.healthplanet.jp/apis/api.html)
- [Fitbit API](https://dev.fitbit.com/build/reference)
