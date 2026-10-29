# kasp-trainee

REST-сервис обмена информацией об угрозах. Контракт: `backend/api/openapi.yaml`.

## Стек

Go, Gin, Postgres, GORM, goose, Docker Compose.

## Быстрый старт

```bash
cp backend/.env.example backend/.env
make up
```

API: `http://localhost:8080`.

```bash
curl -X POST http://localhost:8080/info \
  -H "Content-Type: application/json" \
  -d '{"md5":"d41d8cd98f00b204e9800998ecf8427e","name":"Trojan.Win32.Generic","severity":"high"}'

curl http://localhost:8080/info/d41d8cd98f00b204e9800998ecf8427e
```

## Команды

| target | смысл |
|---|---|
| `up` / `down` | поднять / снести стек |
| `rebuild` | пересобрать backend |
| `test` | юниты (`internal/...`) |
| `test-integration` | `tests/integration` (repo-тест скипнется без `TEST_DATABASE_URL`) |
| `migrate` | goose up локально |
| `fmt` / `lint` | формат / линт |

## Структура

```
backend/
  cmd/api
  api/openapi.yaml
  tests/integration
  internal/
    api/          router, handlers, middlewares, httperr
    services/
    repository/   interface + postgres
    dto/ validation/ config/ logger/ bootstrap/
  migrations/
```

## Почему так сделано

- **Слои handler → service → repository.** HTTP и БД не смешиваются; проще тестировать сервис на моке репо.
- **Модель шире спеки.** К `md5`/`name` добавлены `severity`, `description`, `category` — это как раз «осмысленное дополнение полей» из ТЗ. В OpenAPI они тоже отражены.
- **MD5 в lower-case.** Иначе Postgres посчитает `AbC...` и `abc...` разными уникальными значениями.
- **Ошибки репо → HTTP через `httperr`.** `ErrNotFound`/`ErrConflict` мапятся в 404/409 в одном месте, хендлеры не раздуваются.
- **Middleware:** `X-Trace-Id` (свой или сгенерированный) → access-log со `trace_id` → rate limit по IP (`golang.org/x/time/rate`) → request timeout. Лимиты в конфиге: `rate_limit_rps` / `rate_limit_burst`.
- **Docker multi-stage + отдельный migrate.** Runtime без компилятора; миграции гоняются goose-контейнером до старта backend.
- **Тесты:** юниты рядом с пакетами; интеграционки вынесены в `backend/tests/integration` (HTTP на in-memory репо + опционально живой Postgres).
