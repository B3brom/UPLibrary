# Магазин одежды

## Описание проекта

REST API + встроенный веб-интерфейс для магазина одежды.

**Стек:**
- Go 1.26+
- `github.com/go-chi/chi/v5` — маршрутизация
- `github.com/jmoiron/sqlx` + `modernc.org/sqlite` — БД (чистый Go, без cgo)
- `github.com/golang-jwt/jwt/v5` — авторизация (JWT, HS256)
- `golang.org/x/crypto/bcrypt` — хеширование паролей
- `github.com/joho/godotenv` — конфигурация из `.env`
- `log/slog` — структурированное JSON-логирование
- `embed` — встроенные в бинарник `html/css/js`

**Сущности:** Product, Photo, Role, User, Cart, Order, OrderItem.

**Роли:**
| ID (UUID) | Название |
|---|---|
| `…0001` | Гость |
| `…0002` | Клиент |
| `…0003` | Менеджер |
| `…0004` | Администратор |

**Статусы заказа:** `Принят`, `В работе`, `Собран`, `Готов к выдаче`, `Отменён`.

## Структура проекта
C:.
│   .env
│   exm.sql
│   go.mod
│   go.sum
│   guide.md
│   login.json
│   package.json
│   products.json
│   
├───cmd
│   │   main.go
│   │   
│   └───hash
│           hash.go
│           
├───data
│           
├───internal
│   ├───config
│   │       config.go
│   │       
│   ├───db
│   │       db.go
│   │       
│   ├───handler
│   │       auth.go
│   │       auth_test.go
│   │       cart.go
│   │       common.go
│   │       order.go
│   │       order_item.go
│   │       photo.go
│   │       product.go
│   │       role.go
│   │       user.go
│   │       
│   ├───logger
│   │       logger.go
│   │       
│   ├───middleware
│   │       middleware.go
│   │       
│   ├───models
│   │       models.go
│   │       
│   ├───repository
│   │       cart.go
│   │       order.go
│   │       order_item.go
│   │       photo.go
│   │       product.go
│   │       role.go
│   │       user.go
│   │       
│   ├───router
│   │       router.go
│   │       
│   └───web
│       │   web.go
│       │   
│       └───static
│               app.js
│               index.html
│               style.css
│               
├───migrations
│       001_init.sql
│       embed.go
──────────────────────────

## Запуск

### 1. Переменные окружения

Скопируйте содержимое `.env.example` в `.env`

### 2. Регистрация админа

(по умолчанию админ будет admin|admin@shop.ru|'хеш пароля secret'...)

go run cmd\hash\hash.go 'Впишите пароль'
#->$2a$10$OlDNuACCo0N/P6XS7.5DNuD.aIXtwVm0nYAUsxoxnikvAmAT7w2Ay
Скопируйте поулученный хеш и вставьте его в migrations\001_init.sql\ в инструкцию INSERT OR IGNORE INTO users (id, name, email, phone, password_hash, role_id) VALUES ('550e8400-e29b-41d4-a716-446655440001', 
'Admin', 'admin@shop.ru', '+70000000001', 'ХешСюда', '00000000-0000-0000-0000-000000000004');

### 3. Запуск

go run cmd\main.go

# Или

GOOS=windows GOARCH=amd64 go build -o app.exe
Запустить main.exe

В любом случае откроется консоль и сайт в браузере. Адрес: http://localhost:8080

### Примеры curl-запросов
# Список товаров с фильтрами и сортировкой
curl "http://localhost:8080/product?gender=M&brand=Nike&sort_price=asc&limit=10&offset=0"

# Мультивыбор сезонов (несколько параметров season)
curl "http://localhost:8080/product?season=Осень&season=Демисезон"

# Один товар
curl "http://localhost:8080/product/<PRODUCT_ID>"

# Фотография по id
curl -o photo.jpg "http://localhost:8080/photo/<PHOTO_ID>"

# Список фото товара (для галереи)
curl "http://localhost:8080/photo?product_id=<PRODUCT_ID>"
# Регистрация (возвращает token и объект user)
curl -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{
        "name":"Ivan",
        "email":"ivan@example.com",
        "phone":"+79990000002",
        "password":"password123"
      }'
## Авторизация
# Вход
curl -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@shop.ru","password":"secret"}'
Сохраните токен в переменную для удобства
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@shop.ru","password":"secret"}' | jq -r .token)

# Текущий пользователь
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/me

# Пополнить свой баланс
curl -X POST http://localhost:8080/balance/topup \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"amount": 5000}'

# Пополнить баланс другому пользователю (только админ)
curl -X POST http://localhost:8080/user/<USER_ID>/balance \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"amount": 1000}'
# Добавить (при повторе — инкремент count)
curl -X POST http://localhost:8080/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"<PRODUCT_ID>","count":2}'

# Своя корзина
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/cart/<USER_ID>

# Изменить количество
curl -X PUT http://localhost:8080/cart/<CART_ID> \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"count": 5}'

# Удалить позицию
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/cart/<CART_ID>
# Оформить (цены и наличие берутся из БД, баланс списывается)
curl -X POST http://localhost:8080/order \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"items":[
        {"product_id":"<PRODUCT_ID_1>","quantity":2},
        {"product_id":"<PRODUCT_ID_2>","quantity":1}
      ]}'

# Мои заказы
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/my/orders?limit=20&offset=0"

# Детали одного заказа (клиент видит только свои; менеджер/админ — любые)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/order/<ORDER_ID>

# Позиции заказа
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/order_item/<ORDER_ID>

# Отменить (возврат товара + денег на баланс; доступно владельцу и менеджеру/админу)
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/order/<ORDER_ID>/cancel
# Все заказы с фильтрами
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/order?status=Принят&from=2025-01-01T00:00:00Z&limit=50"

# Сменить статус (Отменён через этот метод недоступен — только через /cancel)
curl -X PUT http://localhost:8080/order/<ORDER_ID> \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"В работе"}'
# Создать товар (при совпадении name+brand+size+color количество суммируется)
curl -X POST http://localhost:8080/product \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
        "name":"Футболка Basic",
        "type":"Футболка",
        "gender":"M",
        "size":"L",
        "color":"Чёрный",
        "description":"100% хлопок",
        "brand":"Nike",
        "season":"Лето",
        "price":1990,
        "quantity":10
      }'

# Обновить
curl -X PUT http://localhost:8080/product/<PRODUCT_ID> \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Футболка Basic","brand":"Nike","price":2190,"quantity":7,"gender":"M"}'

# Удалить
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/product/<PRODUCT_ID>

# Импорт массива товаров из JSON
curl -X POST http://localhost:8080/product/import \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '[
        {"name":"Худи","brand":"H&M","gender":"U","size":"XL","season":"Осень","price":3490,"quantity":5},
        {"name":"Рубашка","brand":"Zara","gender":"M","size":"M","season":"Демисезон","price":2790,"quantity":3}
      ]'

# Загрузить фото товара
curl -X POST http://localhost:8080/photo \
  -H "Authorization: Bearer $TOKEN" \
  -F "product_id=<PRODUCT_ID>" \
  -F "file=@./photo.jpg"

# Удалить фото
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/photo/<PHOTO_ID>

  curl -X POST http://localhost:8080/role \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Курьер"}'

curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/role
curl -X PUT http://localhost:8080/role/<ROLE_ID> \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Курьер 2"}'
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8080/role/<ROLE_ID>
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/user?limit=50&offset=0"
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/user/role/<ROLE_ID>"

# Сменить роль
curl -X PUT http://localhost:8080/user/<USER_ID>/role \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"role_id":"00000000-0000-0000-0000-000000000003"}'

# Удалить
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8080/user/<USER_ID>

# Реализовано

SQLite через database/sql/sqlx

Конфигурация через .env

Структурированное логирование log/slog в формате JSON

Структура

Встраивание статики через //go:embed — один .exe без внешних файлов

Миграция схемы + сид ролей

Сущности и методы (по ТЗ)

Product: POST / GET / GET/{id} / PUT / DELETE, пагинация, фильтрация
(тип, пол, размер, цвет, бренд, сезон), сортировка по цене

Photo: POST / GET /{id} / DELETE, дополнительно GET /photo?product_id=…

Role: POST / GET / PUT / DELETE

User: POST / GET / GET /role/{id} / PUT / DELETE, пагинация

Cart: POST / GET /{user_id} / PUT / DELETE

Order: POST / GET / GET/{id} / PUT, пагинация, фильтрация по времени и статусу

OrderItem: POST / GET /{order_id}

Роли и разграничение доступа

JWT-авторизация

Гость: просмотр каталога и карточки товара

Клиент: корзина, оформление, свои заказы, отмена

Менеджер: все заказы, смена статусов, отмена, CRUD товаров и фото

Администратор: всё выше + управление ролями и пользователями

Сверх ТЗ (в рамках итераций)

UUID как первичные ключи (в ТЗ)

Поле products.quantity, автоматическое суммирование партий

Баланс пользователя, пополнение, оплата заказа с баланса

Отмена заказа с атомарным возвратом товара на склад и денег на баланс

Запрет менять отменённый заказ и выставлять Отменён через PUT /order/{id}

Импорт массива товаров из JSON

Галерея фото, главное фото в каталоге, миниатюры в деталях заказа

Корзина: апсерт по (user_id, product_id), изменение количества в UI

Веб-интерфейс

# Не реализовано

 Загрузка файлов: не ограничены форматы фото и размер.

 Баланс и оплата — оплата производится мнимыми деньгами без интеграции с платежными сервисами.

 Восстановление пароля, подтверждение e-mail, капча — не реализованы.

 Докер — проект реализован на go build, поддержка других устройств не проверена
 
Автоматизированные тесты - на данный момент тестирование ручное

PostgreSQL — отсутствует. Вместо него sqlite

Рукописный код веб части - на данный момент сайт сделан при помощи ИИ на основе "http-ядра", рукописного HTML+CSS скелета

Поддержка мобильных устройств - На данный момент дизайн сайта адаптирован только под ПК