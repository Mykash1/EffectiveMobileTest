# Subscriptions Aggregation Service

REST-сервис для агрегации данных об онлайн подписках пользователей.

## Требования

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 16

## Запуск

```bash
docker compose up --build
```

Сервис будет доступен на `http://localhost:8080`

## Swagger документация

После запуска сервиса:
- Swagger UI: `http://localhost:8080/swagger/`
- OpenAPI JSON: `http://localhost:8080/docs/swagger.json`

## API Endpoints

| Метод  | Путь                    | Описание                                    |
|--------|-------------------------|---------------------------------------------|
| POST   | /subscriptions          | Создать подписку                            |
| GET    | /subscriptions          | Получить список всех подписок               |
| GET    | /subscriptions/{id}     | Получить подписку по ID                     |
| PUT    | /subscriptions/{id}     | Обновить подписку                           |
| DELETE | /subscriptions/{id}     | Удалить подписку                            |
| GET    | /subscriptions/total    | Подсчитать суммарную стоимость подписок     |

### Пример создания подписки

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

### Пример подсчета суммарной стоимости

```bash
curl "http://localhost:8080/subscriptions/total?from=01-2025&to=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"
```

## Конфигурация

Конфигурационные данные вынесены в `.env` файл:

| Переменная    | Описание              | По умолчанию  |
|---------------|-----------------------|---------------|
| APP_PORT      | Порт сервера          | 8080          |
| DB_HOST       | Хост PostgreSQL       | db            |
| DB_PORT       | Порт PostgreSQL       | 5432          |
| DB_USER       | Пользователь БД      | postgres      |
| DB_PASSWORD   | Пароль БД             | postgres      |
| DB_NAME       | Имя базы данных       | subscriptions|
| DB_SSLMODE    | Режим SSL             | disable       |

## Миграции

Миграции автоматически применяются при первом запуске через `docker-entrypoint-initdb.d`.
