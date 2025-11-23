# Pull Request Reviewer Assignment Service

Микросервис для автоматического назначения ревьюеров на Pull Request'ы, разработан как тестовое задание для стажировки в Авито 2025

## Описание

Сервис предоставляет HTTP API для управления командами разработчиков, пользователями и Pull Request'ами. Основная функциональность - автоматическое назначение ревьюеров при создании PR на основе принадлежности к команде и статуса активности участников

## Основные бизнес-правила

### Назначение ревьюеров при создании PR

При создании Pull Request автоматически назначается до 2 ревьюеров из команды автора. Правила выбора:

- Кандидатами могут быть только активные пользователи (is_active = true)
- Автор PR исключается из списка кандидатов
- Если в команде меньше 2 доступных участников, назначается столько, сколько есть (может быть 0, 1 или 2)
- Выбор ревьюеров случайный среди доступных кандидатов

### Переназначение ревьювера

При переназначении одного ревьювера на другого:

- Новый ревьювер выбирается из команды заменяемого ревьювера (не из команды автора)
- Исключаются: автор PR и все текущие ревьюеры
- Замена возможна только для PR в статусе OPEN
- Если нет доступных кандидатов, возвращается ошибка NO_CANDIDATE

### Merge Pull Request

Операция merge переводит PR в статус MERGED и фиксирует время merge:

- Операция идемпотентна: повторный вызов для уже merged PR возвращает 200 с текущим состоянием
- После merge любые изменения ревьюеров запрещены

### Управление командами и пользователями

- Команда создается вместе с участниками через POST /team/add
- Если пользователь уже существует, его данные обновляются (username, team_name, is_active)
- Один пользователь принадлежит только одной команде

## Технологический стек

- Go 1.25
- PostgreSQL 15
- pgx/v5 (драйвер БД)
- chi v5 (HTTP router)
- golang-migrate (миграции)
- Docker, Docker Compose

## Структура проекта

```
Pull_Request_service/
├── cmd/
│   └── api/
│       └── main.go                  # Точка входа приложения
├── internal/
│   ├── domain/                      # Доменные модели и бизнес-логика
│   │   ├── user.go
│   │   ├── team.go
│   │   ├── pull_request.go
│   │   ├── status.go
│   │   └── errors.go
│   ├── repository/                  # Работа с базой данных
│   │   ├── interfaces.go
│   │   ├── postgres.go
│   │   ├── team_repo.go
│   │   ├── user_repo.go
│   │   └── pr_repo.go
│   ├── service/                     # Бизнес логика и оркестрация
│   │   ├── team_service.go
│   │   ├── user_service.go
│   │   └── pr_service.go
│   ├── handler/                     # HTTP обработчики
│   │   ├── router.go
│   │   ├── dto.go
│   │   ├── error.go
│   │   ├── team_handler.go
│   │   ├── user_handler.go
│   │   └── pr_handler.go
│   ├── middleware/                  # HTTP middleware
│   │   ├── logging.go
│   │   └── recovery.go
│   └── config/
│       └── config.go                # Конфигурация из переменных окружения
├── pkg/
│   └── database/
│       └── postgres.go              # Инициализация пула соединений
├── migrations/                      # SQL миграции
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── .env.example
└── README.md
```

### Описание слоев

- **domain** - доменные сущности (User, Team, PullRequest), валидация, доменные ошибки
- **repository** - взаимодействие с PostgreSQL, интерфейсы и реализации
- **service** - бизнес-логика, транзакции, вызовы репозиториев
- **handler** - HTTP API, маппинг между DTO и доменными моделями
- **middleware** - логирование запросов и обработка паник
- **config** - загрузка конфигурации из переменных окружения

## Запуск проекта

### Через Docker Compose (рекомендуется)

Требования: Docker и Docker Compose.

```bash
# Клонировать репозиторий
git clone https://github.com/mivihan/Pull_Request_service.git
cd Pull_Request_service

# Запустить все сервисы (БД, миграции, API)
docker-compose up --build
```

Сервис будет доступен на http://localhost:8080

Проверка работы:

```bash
curl http://localhost:8080/health
```

Остановка:

```bash
docker-compose down         # Остановить сервисы
docker-compose down -v      # Остановить и удалить данные БД
```

### Локальный запуск

Требования: Go 1.25+, PostgreSQL 15+, golang-migrate

Создать базу данных:

```bash
createdb pr_reviewer
```

Применить миграции:

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/pr_reviewer?sslmode=disable" up
```

Настроить переменные окружения:

```bash
cp .env.example .env
# Отредактировать .env при необходимости
```

Запустить сервис:

```bash
go run ./cmd/api
```

Или через Makefile:

```bash
make run
```

## API Endpoints

### Teams

**POST /team/add** - создать команду с участниками

Если пользователь уже существует, его данные обновляются.

Пример запроса:

```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "is_active": true}
    ]
  }'
```

Ответ (201):

```json
{
  "team": {
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "is_active": true}
    ]
  }
}
```

**GET /team/get?team_name=X** - получить команду с участниками

**POST /team/deactivateUsers** – массово деактивировать пользователей команды и пересчитать ревьюверов открытых PR

Тело запроса:

```json
{
  "team_name": "backend",
  "user_ids": ["u2", "u3"]
}
```
Поведение:

    всем указанным пользователям в данной команде устанавливается is_active = false;
    для всех открытых PR, где эти пользователи были ревьюверами:
        если в команде есть другие активные кандидаты (не автор и не текущие ревьюверы), то они назначаются вместо деактивированных;
        если кандидатов нет, то деактивированные ревьюверы просто снимаются, PR остается с меньшим числом ревьюверов

Ответ (200):
```
{
  "team_name": "backend",
  "deactivated_count": 2,
  "affected_pr_count": 3
}
```

### Users

**POST /users/setIsActive** - установить флаг активности пользователя

Пример запроса:

```bash
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "u2",
    "is_active": false
  }'
```

**GET /users/getReview?user_id=X** - получить список PR, где пользователь назначен ревьювером

### Pull Requests

**POST /pullRequest/create** - создать PR с автоматическим назначением ревьюеров

Пример запроса:

```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add authentication",
    "author_id": "u1"
  }'
```

Ответ (201):

```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add authentication",
    "author_id": "u1",
    "status": "OPEN",
    "assigned_reviewers": ["u2", "u3"],
    "createdAt": "2025-01-15T10:30:00Z",
    "mergedAt": null
  }
}
```

**POST /pullRequest/merge** - пометить PR как merged (идемпотентная операция)

Пример запроса:

```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001"
  }'
```

Ответ (200):

```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add authentication",
    "author_id": "u1",
    "status": "MERGED",
    "assigned_reviewers": ["u2", "u3"],
    "createdAt": "2025-01-15T10:30:00Z",
    "mergedAt": "2025-01-15T11:00:00Z"
  }
}
```

**POST /pullRequest/reassign** - переназначить ревьювера

Пример запроса:

```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u2"
  }'
```

Ответ (200):

```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add authentication",
    "author_id": "u1",
    "status": "OPEN",
    "assigned_reviewers": ["u4", "u3"],
    "createdAt": "2025-01-15T10:30:00Z",
    "mergedAt": null
  },
  "replaced_by": "u4"
}
```

### Health

**GET /health** - проверка состояния сервиса

Возвращает 200 OK если сервис работает.

## Формат ошибок

Все ошибки возвращаются в едином формате:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Описание ошибки"
  }
}
```


### Stats

**GET /stats/reviewers** - статистика назначений ревьюеров

Возвращает количество назначений для каждого пользователя, который когда-либо был ревьювером

Пример запроса:

```bash
curl http://localhost:8080/stats/reviewers
```

Ответ(200):
```
{
  "reviewers": [
    {
      "user_id": "u2",
      "assignments_count": 5
    },
    {
      "user_id": "u3",
      "assignments_count": 3
    }
  ]
}
```

**GET /stats/pullRequests** - статистика по статусам PR

Возвращает количество PR в статусах OPEN и MERGED

Пример запроса:

```
curl http://localhost:8080/stats/pullRequests
```

Ответ(200):
```
{
  "open": 7,
  "merged": 3
}
```



### Коды ошибок

- **TEAM_EXISTS** (400) - команда с таким именем уже существует
- **PR_EXISTS** (409) - Pull Request с таким ID уже существует
- **PR_MERGED** (409) - нельзя изменять ревьюеров у PR в статусе MERGED
- **NOT_ASSIGNED** (409) - указанный пользователь не назначен ревьювером этого PR
- **NO_CANDIDATE** (409) - нет доступных активных кандидатов для переназначения
- **NOT_FOUND** (404) - запрашиваемый ресурс не найден (team, user или PR)
- **INVALID_REQUEST** (400) - невалидный формат запроса или отсутствуют обязательные поля
- **INTERNAL_ERROR** (500) - внутренняя ошибка сервера

## Архитектура

Проект следует принципам Clean Architecture с разделением на слои:

### Слои приложения

**HTTP Layer (handler)** - принимает запросы, валидирует входные данные, вызывает сервисы, формирует HTTP ответы

**Service Layer (service)** - реализует бизнес-логику, оркестрирует вызовы репозиториев, управляет транзакциями

**Repository Layer (repository)** - инкапсулирует работу с базой данных, выполняет SQL запросы через pgx/v5

**Domain Layer (domain)** - содержит доменные модели, валидацию, бизнес-правила и типизированные ошибки

### Транзакции

Операции, требующие консистентности данных, выполняются в транзакциях:

- Создание команды с одновременным созданием/обновлением пользователей
- Создание PR с назначением ревьюеров
- Переназначение ревьювера

Транзакции управляются через метод `Repositories.WithTx()`, который использует контекст для передачи транзакции между вызовами репозиториев

### Идемпотентность

Операция merge реализована идемпотентно: повторный вызов для уже merged PR не вызывает ошибку, а возвращает текущее состояние с кодом 200 Это достигается проверкой в методе `PullRequest.Merge()` и на уровне сервиса

## База данных

### Схема

Используется PostgreSQL с четырьмя таблицами:

- **teams** - команды разработчиков
- **users** - пользователи с привязкой к команде
- **pull_requests** - Pull Request'ы
- **pr_reviewers** - связь между PR и назначенными ревьюерами (many-to-many)

### Миграции

Миграции применяются автоматически при запуске через docker-compose. При локальном запуске используйте golang-migrate:

```bash
# Применить миграции
migrate -path migrations -database "$DATABASE_URL" up

# Откатить последнюю миграцию
migrate -path migrations -database "$DATABASE_URL" down 1
```

## Конфигурация

Сервис конфигурируется через переменные окружения. Пример в файле .env.example:

```
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/pr_reviewer?sslmode=disable
LOG_LEVEL=info
```

Переменные:

- **PORT** - порт HTTP сервера (по умолчанию 8080)
- **DATABASE_URL** - строка подключения к PostgreSQL (обязательная)
- **LOG_LEVEL** - уровень логирования: debug, info, warn, error (по умолчанию info)

## Тестирование

Запуск тестов:

```bash
go test -v -race -cover ./...
```

Или через Makefile:

```bash
make test
```

Тесты включают:

- Юнит-тесты доменной логики (internal/domain/pull_request_test.go)
- Тесты бизнес-логики сервисного слоя (internal/service/pr_service_test.go)
- Интеграционные тесты(E2E)

Интеграционные (E2E) тесты находятся в `test/integration/api_test.go` и выполняются поверх запущенного через docker-compose сервиса

Запуск:

```bash
make test-integration

# или вручную:
docker-compose up -d
go test -v ./test/integration/...
docker-compose down
```

Тесты проверяют:

    /health;
    создание команды и PR, автоназначение ревьюверов;
    идемпотентность merge;
    эндпоинты статистики /stats/*;
    поведение /team/deactivateUsers (деактивация и перераспределение ревьюверов)

## Makefile команды

Доступные команды:

- `make build` - собрать бинарник
- `make run` - запустить локально
- `make test` - запустить тесты
- `make docker-up` - запустить через docker-compose
- `make docker-down` - остановить docker-compose
- `make docker-logs` - посмотреть логи API
- `make clean` - удалить build артефакты

## Допущения и ограничения

1. Идентификаторы пользователей, команд и PR передаются извне как строки (не автоинкремент)
2. Пользователь принадлежит только одной команде (связь many-to-one, не many-to-many)
3. При создании команды существующие пользователи обновляются (username, team_name, is_active)
4. Выбор ревьюеров случайный среди доступных кандидатов
5. Миграции применяются автоматически через отдельный Docker контейнер при запуске docker-compose
6. Graceful shutdown реализован с таймаутом 10 секунд

**Нагрузочное тестирование**

Сервис протестирован с помощью k6 для проверки соответствия требованиям по производительности
Требования из ТЗ

    Объем данных: до 20 команд, до 200 пользователей
    RPS: до 5
    SLI времени ответа: 300 мс (95-й перцентиль)
    SLI успешности: 99.9%

Запуск нагрузочного теста

Убедитесь, что сервис запущен:

```Bash
docker-compose up -d
```
Установите k6 (если еще не установлен):
```
# macOS
brew install k6

# Linux/Windows — см. https://k6.io/docs/get-started/installation/
```
Запустите тесты:

```Bash
k6 run test/loadtest.js
```

### Результаты нагрузочного тестирования

Тестирование выполнялось с помощью k6 (сценарии smoke/baseline/stress) против запущенного сервиса на `http://localhost:8080` (docker-compose)

#### Итоговые показатели

| Метрика                | Значение   | Требование (ТЗ)      |
|------------------------|-----------:|----------------------|
| Число запросов         | 17 600     | -                    |
| Средний RPS            | 53.1       | ≈ 5                  |
| Успешность             | 100 %      | >= 99.9 %             |
| p95 latency            | 4.15 ms    | < 300 ms             |
| p99 latency            | ~5 ms      | -                    |
| Среднее время ответа   | 1.78 ms    | -                   |
| Максимальное время     | 24.86 ms   | -                    |

Даже при нагрузке порядка 50 RPS, что существенно выше указанных в задании 5 RPS, сервис укладывается в SLI по задержке и успешности

#### Краткие выводы

- При целевой и повышенной нагрузке (до 30 виртуальных пользователей) 95 % запросов обрабатываются быстрее 4.15 мс, максимальное время ответа остаётся значительно ниже 300 мс
- Ошибок на уровне HTTP и проверок (checks) не зафиксировано: 17 600 успешных запросов из 17 600
- При текущих объёмах данных (до 20 команд и до 200 пользователей) узких мест не выявлено

При дальнейшем масштабировании потенциальными направлениями оптимизации могут быть:
- кеширование агрегатов для эндпоинтов статистики (/stats/*);
- дополнительный анализ запросов к БД в профилировщике и, при необходимости, добавление индексов

## Разработка

Форматирование кода:

```bash
go fmt ./...
```

Линтинг (требуется golangci-lint):

```bash
# линтинг с учётом конфигурации .golangci.yml
make lint
```
Конфигурация линтера описана в файле `.golangci.yml`

## Лицензия

Тестовое задание для стажировки в Авито 2025

GitHub: https://github.com/mivihan/Pull_Request_service