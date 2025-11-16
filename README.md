## PR Review Service

Сервис для управления командами, пользователями и назначениями ревьюеров на Pull Request'ы. Реализованы:
* Создание команды с участниками
* Создание PR с автоматическим назначением активных ревьюеров (до 2)
* Идемпотентный merge PR
* Переназначение ревьюера (reassign) с проверками статуса и кандидатов
* Получение списка PR, где пользователь является ревьюером
* Массовая деактивация пользователей команды с безопасным перераспределением открытых PR
* Статистика использования (общие метрики, топ ревьюеры, метрики по PR)

## Стек и требования
Язык: Go. \
База данных: PostgreSQL\
Спецификация API: `openapi.yml`.

Целевые нефункциональные требования из условия:
* Объём: ≤ 20 команд, ≤ 200 пользователей
* RPS: ~5
* SLI p95 времени ответа: ≤ 300 мс
* SLI успешности: ≥ 99.9%
* Пользователь с `isActive = false` не назначается ревьюером
* Операция merge идемпотентна
* Сервис доступен на `:8080` через `docker-compose up`

## Эндпоинты
| Method | Path | Описание |
|--------|------|----------|
| POST | /team/add | Создать/обновить команду |
| GET  | /team/get?team_name=NAME | Получить команду |
| POST | /pullRequest/create | Создать PR и назначить ревьюеров |
| POST | /pullRequest/merge | Идемпотентный merge |
| POST | /pullRequest/reassign | Переназначить ревьюера |
| POST | /users/setIsActive | Сменить активность пользователя |
| GET  | /users/getReview?user_id=ID | PR где пользователь ревьюер |
| POST | /users/massDeactivate | Массовая деактивация + перераспределение |
| GET  | /stats/get | Общая статистика |

Полное описание — см. `openapi.yml`.

## Установка и запуск

Предварительно установите Docker + docker-compose.

```bash
make docker-up
```

## Дополнительные задачи (Checklist)

- [x] Простой эндпоинт статистики (`GET /stats/get`) — возвращает агрегаты: количество назначений по пользователям (топ ревьюеры), общее количество PR, распределение состояний.
- [x] Нагрузочное / стресс-тестирование — проведено с помощью k6, результаты ниже.
- [x] Массовая деактивация пользователей команды (`POST /users/massDeactivate`) с безопасной переназначаемостью открытых PR (исключаются уже назначенные, избегая UNIQUE конфликтов). Среднее время ответа под нагрузкой укладывается в < 100 мс.
- [x] Интеграционные тесты (`internal/storage/*_test.go`, `internal/handlers/*_test.go`) и E2E-подобные проверки через `httptest`.
- [x] Конфигурация линтера — `make lint` запускает `golangci-lint`. Используется стандартная конфигурация `.golangci.yaml`

## Результаты стресс-тестирования (k6)

Итоговая сводка:

```text
█ THRESHOLDS 

	http_req_duration
	✓ 'p(95)<350' p(95)=21.95ms

	http_req_failed
	✓ 'rate<0.005' rate=0.00%


█ TOTAL RESULTS 

	checks_total.......: 22284   37.12723/s
	checks_succeeded...: 100.00% 22284 out of 22284
	checks_failed......: 0.00%   0 out of 22284

	✓ status 2xx
	✓ response time <350ms
	✓ no server errors

	HTTP
	http_req_duration..............: avg=8.12ms   min=497µs    med=5.39ms   max=177.12ms p(90)=16.76ms  p(95)=21.95ms 
		{ expected_response:true }...: avg=8.12ms   min=497µs    med=5.39ms   max=177.12ms p(90)=16.76ms  p(95)=21.95ms 
	http_req_failed................: 0.00%  0 out of 7439
	http_reqs......................: 7439   12.39407/s

	EXECUTION
	iteration_duration.............: avg=509.23ms min=500.72ms med=506.53ms max=678.31ms p(90)=517.92ms p(95)=523.21ms
	iterations.....................: 7428   12.375743/s
	vus............................: 1      min=1         max=12
	vus_max........................: 12     min=12        max=12

	NETWORK
	data_received..................: 156 MB 260 kB/s
	data_sent......................: 1.2 MB 2.0 kB/s

running (10m00.2s), 00/12 VUs, 7428 complete and 0 interrupted iterations
default ✓ [======================================] 00/12 VUs  10m0s
```

## Команды Makefile

Сборка и проверка:
```bash
make deps        # Загрузка зависимостей
make build       # Сборка бинарника
make run         # Запуск сервера (8080)
make lint        # Запуск golangci-lint
```

Docker:
```bash
make docker-up           # Поднять сервисы
make docker-down         # Остановить
make docker-down-clean   # Полная очистка (volumes, images)
make docker-rebuild      # Пересборка контейнеров
make docker-logs         # Все логи
make docker-logs-app     # Логи приложения
make docker-logs-db      # Логи DB
make status              # Статус контейнеров
```

База данных:
```bash
make db-clean    # TRUNCATE таблиц
make db-reset    # Полный reset окружения
```

Тесты и нагрузка:
```bash
go test ./...          # Все тесты
make stress-test       # Стресс-тест k6 (предварительная очистка)
make load-test         # Сценарий нагрузки
make load-test-fixed   # Сценарий с подготовленными данными
make load-test-summary # Быстрый тест с summary
make load-test-docker  # Запуск k6 внутри Docker
make benchmark         # Apache Bench (простая метрика)
```

Прочее:
```bash
make tidy      # go mod tidy
make verify    # go mod verify
make init-loadtest # создать директорию scripts
```



