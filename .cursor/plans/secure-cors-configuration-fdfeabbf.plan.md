<!-- fdfeabbf-881a-42c4-bc01-a5df97cfa553 dc83b873-01e1-41bd-8f29-f27d1b042e99 -->
# План: Модель данных для заявок на закрытие нарушений

## Текущая ситуация

- Существует модель `Violation` с полями: ID, UserID, Type, Description, Status, Photos, CreatedAt
- Статусы: `new`, `confirmed`, `resolved`
- Нет модели для заявок на закрытие

## Требования

- Отдельное поле `closure_type` в заявке (`partial`/`full`)
- Несколько заявок на полное закрытие (история всех попыток)
- Статус Violation меняется при создании заявки (не дожидаясь модерации)
- Отдельное поле `created_by_user_id` в заявке (автор заявки на закрытие)

## Предлагаемая модель

### 1. Новая таблица `violation_closure_requests`

Хранит все заявки на закрытие (частичные и полные):

- `id` UUID PRIMARY KEY
- `violation_id` UUID NOT NULL (FK на violations)
- `closure_type` TEXT NOT NULL CHECK (closure_type IN ('partial', 'full'))
- `created_by_user_id` BIGINT NOT NULL (автор заявки на закрытие)
- `comment` TEXT (комментарий к заявке)
- `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
- `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

Индексы:

- `idx_closure_requests_violation_id` на `violation_id`
- `idx_closure_requests_type` на `closure_type`
- `idx_closure_requests_created_at` на `created_at`

### 2. Новая таблица `violation_closure_photos`

Фото для каждой заявки на закрытие:

- `id` UUID PRIMARY KEY
- `closure_request_id` UUID NOT NULL (FK на violation_closure_requests)
- `url` TEXT NOT NULL
- `thumb_url` TEXT
- `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

Индексы:

- `idx_closure_photos_request_id` на `closure_request_id`

### 3. Обновление статусов Violation

Добавить новые статусы в CHECK constraint:

- `status` TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'confirmed', 'resolved', 'partially_resolved'))

Логика изменения статуса:

- При создании заявки с `closure_type='partial'` → статус `partially_resolved`
- При создании заявки с `closure_type='full'` → статус `resolved`

### 4. Обновление модели в Go (`Server/internal/model/model.go`)

Добавить:

- `ViolationClosureRequest` struct с полями: ID, ViolationID, ClosureType, CreatedByUserID, Comment, Photos, CreatedAt, UpdatedAt
- `ViolationClosurePhoto` struct с полями: ID, ClosureRequestID, URL, ThumbnailURL, CreatedAt
- Обновить `Violation` struct: добавить поле `ClosureRequests []ViolationClosureRequest` (опционально, для детального просмотра)

### 5. SQL запросы (`Server/internal/repository_sqlc/query.sql`)

Добавить:

- `CreateClosureRequest` - создание заявки на закрытие
- `GetClosureRequestsByViolationID` - получение всех заявок для нарушения
- `AddClosurePhoto` - добавление фото к заявке
- `GetClosurePhotosByRequestID` - получение фото заявки
- `UpdateViolationStatus` - обновление статуса нарушения

### 6. Репозиторий (`Server/internal/repository/`)

Добавить методы:

- `CreateClosureRequest(ctx, violationID, closureType, userID, comment)` - создание заявки
- `GetClosureRequestsByViolationID(ctx, violationID)` - получение всех заявок
- `AddClosurePhoto(ctx, requestID, url, thumbURL)` - добавление фото
- `GetClosurePhotosByRequestID(ctx, requestID)` - получение фото заявки
- `UpdateViolationStatus(ctx, violationID, status)` - обновление статуса

### 7. Сервис (`Server/internal/service/`)

Добавить методы:

- `CreateClosureRequestWithPhotos(ctx, violationID, closureType, userID, comment, files, maxPhotos, upload)` - создание заявки с фото
- Создает заявку через репозиторий
- Загружает фото через `uploader.Upload()` с ключом `closure_requests/{request_id}/{filename}`
- Сохраняет URL фото в БД через `AddClosurePhoto`
- Обновляет статус нарушения: `partial` → `partially_resolved`, `full` → `resolved`

### 8. HTTP Handler (`Server/internal/app/http/`)

Создать `create_closure_request.go`:

- Handler принимает `POST /violations/{id}/closure-request`
- Content-Type: `multipart/form-data` (обязательно для фото)
- Поля формы:
- `closure_type` (required): `partial` или `full`
- `comment` (optional): текстовый комментарий
- `photos[]` (optional): массив файлов изображений
- Использует `objectstorage.Uploader` для загрузки фото
- Лимит фото: настраиваемый параметр (например, `maxPhotosPerClosureRequest`)
- Возвращает созданную заявку с массивом фото

### 9. Регистрация handler в `Server/internal/app/app.go`

- Добавить маршрут для создания заявки на закрытие
- Передать `uploader` и `maxPhotosPerClosureRequest` в handler

## Файлы для изменения

- `Server/migrations/` - новая миграция для таблиц
- `Server/internal/model/model.go` - новые структуры
- `Server/internal/repository_sqlc/query.sql` - SQL запросы
- `Server/internal/repository/` - методы репозитория