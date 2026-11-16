# Магазин одежды

## Описание проекта

REST API + встроенный веб-интерфейс для магазина одежды

**Стек:**
- Go 1.26
- `github.com/go-chi/chi/v5` — маршрутизация
- `github.com/jmoiron/sqlx` + `modernc.org/sqlite` — БД
- `github.com/golang-jwt/jwt/v5` — JWT-авторизация (HS256)
- `golang.org/x/crypto/bcrypt` — хеширование паролей
- `github.com/joho/godotenv` — конфигурация из `.env`
- `log/slog` — структурированное JSON-логирование
- `embed` — встроенные в бинарник `html/css/js`

**Сущности:** `Product`, `Photo`, `Role`, `User`, `Cart`, `Order`, `OrderItem`.

### Роли

| ID (UUID) | Название |
|---|---|---|
| `00000000-0000-0000-0000-000000000001` | Гость | Только для неавторизованных  |
| `00000000-0000-0000-0000-000000000002` | Клиент | Через регистрацию |
| `00000000-0000-0000-0000-000000000003` | Менеджер | Только админом |
| `00000000-0000-0000-0000-000000000004` | Администратор | Только админом |

### Статусы заказа

Доступные статусы: `Принят`, `В работе`, `Собран`, `Готов к выдаче`, `Получен`, `Отменён`.

**Правила переходов:**
- `Отменён` — только через `POST /order/{id}/cancel` с указанием причины.
  - Клиент может отменить только на этапе `Принят`; менеджер/админ — на любом активном этапе (`Принят`, `В работе`, `Собран`, `Готов к выдаче`).
  - При отмене товар возвращается на склад, деньги — на баланс владельца заказа.
- `Получен` — только менеджер/админ, только из `Готов к выдаче`.
  - После `Получен` заказ нельзя изменить, отменить или вернуть деньги.
- `Отменён` и `Получен` — окончательные статусы, не позволяющие в дальнейшем изменять статус заказа.

---

## Структура проекта

```text
.
├── cmd/
│   ├── main.go          # точка входа сервера
│   └── hash/
│       └── hash.go      # утилита генерации bcrypt-хеша
├── internal/
│   ├── config/          # .env → Config
│   ├── db/              # подключение + миграции
│   ├── handler/         # HTTP-хендлеры
│   ├── logger/          # slog
│   ├── middleware/      # Auth (JWT), RequireRole, Logger
│   ├── models/          # структуры сущностей
│   ├── repository/      # SQL-запросы
│   ├── router/          # маршруты
│   └── web/static/      # Frontend
├── migrations/
│   ├── 001_init.sql     # Миграция
│   └── embed.go
├── data/                # создаётся автоматически: store.db
├── uploads/             # создаётся автоматически: фото товаров
├── .env
├── go.mod
└── README.md
```

---

## Запуск

### 1. Переменные окружения

Скопируйте или переименуйте `.env.example` в `.env`:

```env
SERVER_PORT=8080
DB_PATH=./data/store.db
UPLOAD_DIR=./uploads
JWT_SECRET=secret
JWT_TTL=24h
LOG_LEVEL=info
```

### 2. Зависимости

```bash
go mod tidy
```

### 3. Регистрация администратора

В `migrations/001_init.sql` есть `INSERT OR IGNORE INTO users …` с админом `admin@shop.ru`. Пароль задается хешем в этом `INSERT`. Чтобы сменить пароль:

```bash
go run ./cmd/hash "ваш-пароль"

```

Скопируйте хеш в `migrations/001_init.sql`:

```sql
INSERT OR IGNORE INTO users (id, name, email, phone, password_hash, role_id) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'Admin', 'admin@shop.ru', '+70000000001',
 '\$2a\$10\$…', '00000000-0000-0000-0000-000000000004');
```

### 4. Запуск

**Через `go run`:**

```bash
go run cmd\main.go
```

**Сборка в один `.exe`:**

```bash

go build cmd\main.go

# запуск
main.exe
```

> Рядом с `.exe` положите `.env` (или ничего — тогда данные пойдут в `%AppData%\ClothingStore\`).

При старте сервер сам создаёт `data/store.db` и `uploads/`, выполняет миграцию и открывает браузер на `http://localhost:8080`.

## Примеры curl-запросов

Базовый URL: `http://localhost:8080`.

### Публичные (гость)

```bash
# Список товаров: поиск, фильтры, сортировка, пагинация
curl "http://localhost:8080/product?q=Samurai&gender=U&brand=SAMURAI&size=XL&sort_price=asc"

# Мультивыбор сезонов (несколько параметров season)
curl "http://localhost:8080/product?season=Осень&season=Демисезон"

# Поиск по артикулу (UUID), без учёта регистра
curl "http://localhost:8080/product?q=c9731527-1994-46e6-8227-e31fd9c0eb77"

```

### Авторизация

```bash
# Регистрация — email требует @ и точку, пароль ≥ 8 символов
curl -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{
        "name":"aaa",
        "email":"aaa@mail.ru",
        "phone":"+79990000002",
        "password":"12345678"
      }'
# → { "token": "...", "user": { "id", "name", "email", "role_id", "balance" } }

# Вход
curl -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@shop.ru","password":"secret"}'
```

**Сохранить токен в переменную:**

```bash
TOKEN=\$(curl -s -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@shop.ru","password":"secret"}' | jq -r .token)
```

### Профиль и баланс

```bash
# Текущий пользователь
curl -H "Authorization: Bearer \$TOKEN" http://localhost:8080/me

# Пополнить свой баланс
curl -X POST http://localhost:8080/balance/topup \
  -H "Authorization: Bearer \$TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"amount": 5000}'

# Пополнить чужой баланс (только админ)
curl -X POST http://localhost:8080/user/<USER_ID>/balance \
  -H "Authorization: Bearer \$TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"amount": 1000}'
```

> Лимит суммы пополнения задаётся в `internal/handler/user.go` (сейчас `1_000_000` за раз).

### Корзина (клиент)

```bash
# Добавить (при повторе count инкрементируется через UNIQUE user_id+product_id)
curl -X POST http://localhost:8080/cart \
  -H "Authorization: Bearer \$TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"<PRODUCT_ID>","count":2}'

# Своя корзина
curl -H "Authorization: Bearer \$TOKEN" http://localhost:8080/cart/<USER_ID>

# Изменить количество
curl -X PUT http://localhost:8080/cart/<CART_ID> \
  -H "Authorization: Bearer \$TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"count": 5}'

# Удалить позицию
curl -X DELETE -H "Authorization: Bearer \$TOKEN" \
  http://localhost:8080/cart/<CART_ID>
```

### Оформление заказа (клиент)

Цена не передаётся — берётся из БД. Наличие проверяется, склад списывается, баланс уменьшается — всё атомарно.

```bash
curl -X POST http://localhost:8080/order \
  -H "Authorization: Bearer \$TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"items":[
        {"product_id":"<PRODUCT_ID_1>","quantity":2},
        {"product_id":"<PRODUCT_ID_2>","quantity":1}
      ]}'
```

**Возможные ошибки (`400 Bad Request`):**
- `Недостаточно средств. Требуется X ₽`
- `«Название»: в наличии N шт., запрошено M шт.`

### Заказы — просмотр

```bash
# Свои заказы (краткий список: №, имя, кол-во вещей, сумма, статус)
curl -H "Authorization: Bearer \$TOKEN" "http://localhost:8080/my/orders?limit=20&offset=0"

# Детали одного заказа: UUID заказа и пользователя, позиции с названиями и фото
# Клиент — только свои (иначе 403). Менеджер/админ — любые.
curl -H "Authorization: Bearer \$TOKEN" http://localhost:8080/order/<ORDER_ID>

# Позиции заказа
curl -H "Authorization: Bearer \$TOKEN" http://localhost:8080/order_item/<ORDER_ID>
```

### Отмена заказа (с причиной)

Поле `reason` обязательно (1–255 символов). Клиент может отменить заказ только на этапе `Принят`.

```bash
# Клиент — только свой заказ в статусе «Принят»
curl -X POST http://localhost:8080/order/<ORDER_ID>/cancel \
  -H "Authorization: Bearer \$TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"reason":"Передумал(а) оформлять заказ"}'

# Менеджер/админ — любой активный (Принят/В работе/Собран/Готов к выдаче)
curl -X POST http://localhost:8080/order/<ORDER_ID>/cancel \
  -H "Authorization: Bearer \$TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"reason":"Товар повреждён на складе"}'

# Менеджер/админ — любой активный (Принят/В работе/Собран/Готов к выдаче)
curl -X POST http://localhost:8080/order/<ORDER_ID>/cancel \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"reason":"Товар повреждён на складе"}'
```

---

### Управление заказами (менеджер и админ)

```bash
# Список всех заказов с фильтрацией по статусу и пагинацией
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/order?status=Принят&limit=20&offset=0"

# Смена статуса заказа
# Допустимые переходы: Принят → В работе → Собран → Готов к выдаче → Получен
curl -X PATCH http://localhost:8080/order/<ORDER_ID>/status \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"В работе"}'
```

---

### Управление товарами и фото (менеджер и админ)

```bash
# Создание товара
curl -X POST http://localhost:8080/product \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
        "name": "Худи Oversize Black",
        "description": "Плотный хлопок 420 г/м², футер трёхнитка",
        "brand": "Nike",
        "category": "Толстовки",
        "gender": "M",
        "size": "L",
        "season": "Осень",
        "price": 5490,
        "stock": 15
      }'

# Обновление товара
curl -X PUT http://localhost:8080/product/<PRODUCT_ID> \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
        "name": "Худи Oversize Black (обновл.)",
        "description": "Плотный хлопок 420 г/м²",
        "brand": "Nike",
        "category": "Толстовки",
        "gender": "M",
        "size": "L",
        "season": "Демисезон",
        "price": 4990,
        "stock": 12
      }'

# Удаление товара (и связанных фото с диска)
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/product/<PRODUCT_ID>

# Загрузка фото для товара (multipart/form-data)
curl -X POST http://localhost:8080/photo \
  -H "Authorization: Bearer $TOKEN" \
  -F "product_id=<PRODUCT_ID>" \
  -F "file=@./path/to/hoodie.jpg"

# Удаление фото
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/photo/<PHOTO_ID>
```

---

### Администрирование пользователей и ролей (только админ)

```bash
# Список всех пользователей
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/user?limit=50&offset=0"

# Данные конкретного пользователя
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/user/<USER_ID>

# Смена роли пользователя (например, выдача прав Менеджера)
curl -X PATCH http://localhost:8080/user/<USER_ID>/role \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"role_id":"00000000-0000-0000-0000-000000000003"}'

# Список доступных ролей
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/role
```

---

## Сводная таблица эндпоинтов

| Метод | Маршрут | Доступ | Описание |
|---|---|---|---|
| `GET` | `/product` | Все | Список товаров с поиском, фильтрами и пагинацией |
| `GET` | `/product/{id}` | Все | Карточка товара |
| `POST` | `/product` | Менеджер, Админ | Создать товар |
| `PUT` | `/product/{id}` | Менеджер, Админ | Обновить товар |
| `DELETE` | `/product/{id}` | Менеджер, Админ | Удалить товар |
| `GET` | `/photo` | Все | Список фото товара (`?product_id=...`) |
| `GET` | `/photo/{id}` | Все | Скачать / отобразить файл изображения |
| `POST` | `/photo` | Менеджер, Админ | Загрузить фото товара |
| `DELETE` | `/photo/{id}` | Менеджер, Админ | Удалить фото товара |
| `POST` | `/register` | Гость | Регистрация нового клиента |
| `POST` | `/login` | Гость | Вход и получение JWT-токена |
| `GET` | `/me` | Авторизованные | Профиль текущего пользователя |
| `POST` | `/balance/topup` | Авторизованные | Пополнение собственного баланса |
| `GET` | `/cart/{user_id}` | Клиент (своя) | Просмотр корзины |
| `POST` | `/cart` | Клиент | Добавить товар в корзину |
| `PUT` | `/cart/{id}` | Клиент (своя) | Изменить количество товара в корзине |
| `DELETE` | `/cart/{id}` | Клиент (своя) | Удалить позицию из корзины |
| `POST` | `/order` | Клиент | Оформить заказ из списка позиций |
| `GET` | `/my/orders` | Клиент | Список своих заказов |
| `GET` | `/order` | Менеджер, Админ | Список всех заказов магазина |
| `GET` | `/order/{id}` | Владелец, Менеджер, Админ | Детали заказа |
| `GET` | `/order_item/{id}` | Владелец, Менеджер, Админ | Позиции заказа |
| `PATCH` | `/order/{id}/status` | Менеджер, Админ | Изменить статус заказа |
| `POST` | `/order/{id}/cancel` | Владелец (`Принят`), Менеджер, Админ | Отменить заказ с возвратом средств и остатков |
| `GET` | `/user` | Админ | Список всех пользователей |
| `GET` | `/user/{id}` | Админ | Профиль пользователя |
| `PATCH` | `/user/{id}/role` | Админ | Изменить роль пользователя |
| `POST` | `/user/{id}/balance` | Админ | Пополнить баланс пользователя |
| `GET` | `/role` | Админ | Справочник ролей |

---

## Встроенный веб-интерфейс

Фронтенд расположен в `internal/web/static/` (`index.html`, `style.css`, `app.js`) и компилируется внутрь бинарного файла через стандартный пакет `embed`.

**Возможности интерфейса:**
- Каталог товаров с поиском, фильтрацией по полу, бренду, размеру, сезону и сортировкой по цене.
- Просмотр карточки товара с галереей фотографий.
- Регистрация, вход, хранение JWT в `localStorage`, отображение текущего баланса и его пополнение.
- Полноценная корзина с пересчётом итоговой суммы и оформлением заказа.
- Личный кабинет с историей заказов, просмотром состава заказа и кнопкой отмены (с вводом причины).
- Панель управления для **Менеджера / Админа**: добавление и редактирование товаров, загрузка/удаление фотографий, обработка заказов и смена их статусов.
- Вкладка **Администрирования** (только для роли `Администратор`): просмотр пользователей, выдача прав менеджера/админа и ручное пополнение баланса.

---

## Нереализованный функционал

Не были реализованы следующие возможности:

1. **Refresh-токены и отзыв JWT (Blacklist):**
   - Используется один короткоживущий или суточный access-токен (`JWT_TTL=24h`). При смене роли или блокировке пользователя старый токен продолжает действовать до истечения срока.
2. **Восстановление и смена пароля через UI/API:**
3. **Интеграция с реальными платёжными шлюзами:**
4. **Валидация изображений:**
5. **Отзывы, рейтинги и вишлисты:**
6. **Промокоды и система скидок:**
   - Цена товара фиксирована; нет поддержки купонов, акций или накопительных скидок.
7. **Уведомления (Push):**
   - Пользователь не получает уведомлений при смене статуса заказа.
8. **Автоматизированные тесты:**
   - Бизнес-логика тестировалась вручную через веб-интерфейс и `curl`.
