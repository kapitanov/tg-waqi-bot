# tg-waqi-bot

A telegram bot that provides current air quality status (with updates).
Data is provided by [waqi.info](https://waqi.info/).

This bot is non-public due to waqi.info API restrictions, so you'll need to set up your own instance of this bot to use it.

## How to build and run

1. Clone git repository to any approprivate directory:

   ```shell
   cd /opt
   git clone https://github.com/kapitanov/tg-waqi-bot.git
   cd tg-waqi-bot
   ```

2. Create a `.env` file (see [configuration](#configuration) section below):

   ```env
   WAQI_TOKEN=waqi-api-token
   TELEGRAM_API_TOKEN=telegram-bot-toket
   TELEGRAM_USERNAMES=your-telegram-username-or-id
   ```

   You'll need to:

   * get an access token for api.waqi.info [here](https://aqicn.org/data-platform/token/)
   * get a bot api token for Telegram [here](http://t.me/BotFather)

3. Build and run docker container:

   ```shell
   docker-compose up -d --build
   ```

## Configuration

This bot is configured via env variables:

| Variable              | Default                    | Description                                                      |
| --------------------- | -------------------------- | ---------------------------------------------------------------- |
| `WAQI_URL`            | `https://api.waqi.info/`   | WAQI service root URL                                            |
| `WAQI_TOKEN`          | Required                   | WAQI service access token                                        |
| `WAQI_CACHE_PATH`     | `/var/tg-waqi-bot/cache`   | Path to WAQI service cache                                       |
| `WAQI_CACHE_DURATION` | `15m`                      | WAQI service cache duration                                      |
| `LISTEN_ADDR`         | `0.0.0.0:8000`             | REST API listen address                                          |
| `BOT_DB_PATH`         | `/var/tg-waqi-bot/bot.dat` | PAth to bot DB file                                              |
| `TELEGRAM_API_URL`    | `https://api.telegram.org` | Telegram bot API URL                                             |
| `TELEGRAM_API_TOKEN`  | Required                   | Telegram bot API access token                                    |
| `TELEGRAM_USERNAMES`  | Required                   | List of allowed Telegram usernames (or userIDs), space separated |

## License

[MIT](LICENSE)
