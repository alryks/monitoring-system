# Monitoring System

Распределенная система мониторинга и управления Docker-инфраструктурой. Проект состоит из основного узла `core` и дочерних узлов `agent`: агенты собирают метрики хоста и контейнеров, отправляют их на Core, получают команды и выполняют действия через Docker API и локальную конфигурацию NGINX.

## Возможности

- мониторинг CPU, RAM, дисков, сети и Docker-контейнеров;
- просмотр агентов, контейнеров, образов, логов и дашборда через веб-интерфейс;
- JWT-аутентификация пользователя;
- управление Docker-контейнерами через очередь действий;
- привязка доменов к агентам и генерация NGINX-конфигураций;
- reverse proxy для маршрутизации UI/API и пользовательских доменов;
- базовые уведомления и настройки нотификаций.

## Стек

- Backend: Go, Chi, PostgreSQL, JWT, Swagger
- Frontend: React, TypeScript, Vite, Recharts, Axios
- Agent: Go, Docker API, gopsutil
- Infrastructure: Docker Compose, NGINX, reverse proxy на Go

## Структура проекта

```text
.
├── core/
│   ├── app/              # React-приложение
│   ├── server/           # Go API, БД, бизнес-логика
│   ├── reverse-proxy/    # Go reverse proxy для UI/API и доменов
│   ├── docker-compose.yml
│   └── .env.example
├── agent/
│   ├── main.go           # сбор метрик и обработка действий
│   ├── docker-compose.yml
│   └── .env.example
└── docs/                 # описание задачи, план, DBML, примеры
```

## Архитектура

`Core` является центральным узлом системы. Он поднимает PostgreSQL, API-сервер, веб-интерфейс и reverse proxy. API хранит агентов, метрики, контейнеры, образы, действия и доменные маршруты.

`Agent` запускается на каждой наблюдаемой машине. Он периодически отправляет данные на `POST /api/agent/ping` с Bearer-токеном агента. В ответ Core возвращает список ожидающих действий, например запуск/остановку контейнера, загрузку образа или обновление NGINX-конфига. После выполнения агент отправляет результат действия обратно в Core.

Пользовательский трафик проходит через reverse proxy. Запросы к домену приложения маршрутизируются в UI или API, а остальные домены сопоставляются с IP агентов через публичный endpoint Core.

## Требования

- Docker и Docker Compose
- Go 1.24+ для локального запуска Go-сервисов без Docker
- Node.js 20+ и npm для локального запуска фронтенда
- доступ к Docker socket на узлах, где запускается агент

## Быстрый запуск Core

1. Подготовьте переменные окружения:

```bash
cd core
cp .env.example .env
```

2. Отредактируйте `core/.env`:

```env
USER=admin
PASSWORD=admin

POSTGRES_DB=monitoring
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password

JWT_SECRET=change-me
APP_DOMAIN=localhost

SERVER_URL=http://server:8000
APP_URL=http://app:3000
```

3. Запустите Core:

```bash
docker compose up --build -d
```

4. Откройте веб-интерфейс:

```text
http://localhost
```

По умолчанию сервер создает администратора `admin/admin`, если в базе еще нет пользователя с ролью `admin`.

## Запуск Agent

1. Создайте агента в веб-интерфейсе Core и скопируйте его токен.

2. Подготовьте переменные окружения:

```bash
cd agent
cp .env.example .env
```

3. Отредактируйте `agent/.env`:

```env
URL=http://<core-host>/api/agent/ping
TOKEN=<agent-token>
INTERVAL=5
```

4. Запустите агент:

```bash
docker compose up --build -d
```

Агент работает в `host` network mode, использует Docker socket и монтирует файловую систему хоста для сбора системных метрик. Это требует повышенных прав, поэтому запускайте агент только на доверенных узлах.

## Основные URL

- UI: `http://localhost`
- API: `http://localhost/api`
- Swagger UI: `http://localhost/api/swagger/index.html`
- Health check reverse proxy: `http://localhost/health`
- PostgreSQL на хосте: `localhost:5433`

## Локальная разработка

Backend:

```bash
cd core/server
POSTGRES_URL='postgres://postgres:password@localhost:5433/monitoring?sslmode=disable' \
JWT_SECRET='dev-secret' \
go run .
```

Frontend:

```bash
cd core/app
npm install
npm run dev
```

Agent:

```bash
cd agent
URL=http://localhost/api/agent/ping TOKEN=<agent-token> INTERVAL=5 go run .
```

## Команды

Core:

```bash
cd core
docker compose up --build -d
docker compose logs -f
docker compose down
```

Agent:

```bash
cd agent
docker compose up --build -d
docker compose logs -f agent
docker compose down
```

Frontend:

```bash
cd core/app
npm run build
npm run lint
```

Go-сервисы:

```bash
cd core/server && go test ./...
cd ../../agent && go test ./...
cd ../core/reverse-proxy && go test ./...
```

## Переменные окружения

Core:

| Переменная | Назначение |
| --- | --- |
| `USER` | логин администратора для конфигурации приложения |
| `PASSWORD` | пароль администратора для конфигурации приложения |
| `POSTGRES_DB` | имя базы PostgreSQL |
| `POSTGRES_USER` | пользователь PostgreSQL |
| `POSTGRES_PASSWORD` | пароль PostgreSQL |
| `JWT_SECRET` | секрет для подписи JWT |
| `APP_DOMAIN` | домен панели управления |
| `SERVER_URL` | внутренний URL backend-сервера |
| `APP_URL` | внутренний URL frontend-приложения |

Agent:

| Переменная | Назначение |
| --- | --- |
| `URL` | адрес endpoint `POST /api/agent/ping` |
| `TOKEN` | Bearer-токен агента из Core |
| `INTERVAL` | интервал отправки метрик в секундах |

## API и действия агентов

Core API принимает метрики от агентов, сохраняет состояние инфраструктуры и возвращает ожидающие действия. Поддерживаемые действия на стороне агента:

- `start_container`
- `stop_container`
- `restart_container`
- `remove_container`
- `pull_image`
- `remove_image`
- `create_nginx_config`
- `update_nginx_config`
- `delete_nginx_config`
- `get_nginx_config`

Результат выполнения действия отправляется в `PUT /api/actions/{id}/status`.

## База данных

Сервер сам создает необходимые таблицы при старте. Основные сущности:

- `users`
- `agents`
- `agent_pings`
- `cpu_metrics`, `memory_metrics`, `disk_metrics`, `network_metrics`
- `containers`, `container_logs`
- `images`, `image_tags`
- `actions`
- `domains`, `domain_routes`

Дополнительная схема описана в `docs/db.dbml`.

