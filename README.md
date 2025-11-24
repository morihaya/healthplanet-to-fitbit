# healthplanet-to-fitbit

[![dockeri.co](https://dockeri.co/image/tattsun/healthplanet-to-fitbit)](https://hub.docker.com/r/tattsun/healthplanet-to-fitbit)

[HealthPlanet](https://www.healthplanet.jp/)に登録された体重・体脂肪率情報を Fitbit へ転送する。

## 環境変数

`.env`ファイルまたは、環境変数に以下を定義する。
※ アクセストークンやリフレッシュトークンは、後述のツールを実行することで自動的に設定ファイル（`~/.config/healthplanet-to-fitbit/config.json`）に保存されるため、手動での設定は不要です。

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

設定ファイルから認証情報を読み込み、直近３か月の情報（体重・体脂肪率）が HeathPlanet から取得され、Fitbit へ登録される。
Fitbit のアクセストークンが期限切れの場合は、自動的にリフレッシュされ、設定ファイルが更新される。

繰り返し起動するとアクセス数の制限に引っかかる場合があるため、時間をおいて起動することを推奨する。
