# Изменения в API для мобильного приложения

## Обзор изменений

Мы переработали модель данных для работы с заявками на закрытие нарушений. Теперь все фото хранятся в заявках, а не напрямую в нарушениях.

**Последнее обновление:** Добавлено поле `requests[]` в ответ `GET /api/violations/{id}` для отображения истории всех заявок с датами, авторами, комментариями и фото.

## Основные изменения

### 1. Модель данных

**Раньше:**
- `Violation` содержал поле `photos[]` с массивом фото
- Фото привязывались напрямую к нарушению

**Теперь:**
- `Violation` больше не содержит поле `photos[]`
- Все фото хранятся в заявках (`violation_requests`)
- При создании `Violation` автоматически создается заявка со статусом `open`
- Фото привязываются к заявкам, а не к нарушениям напрямую

### 2. Структура заявки (ViolationRequest)

```json
{
  "id": "uuid",
  "violation_id": "uuid",
  "status": "open" | "partially_closed" | "closed",
  "created_by_user_id": 123,
  "comment": "текст комментария (опционально)",
  "photos": [
    {
      "id": "uuid",
      "request_id": "uuid",
      "url": "https://cdn/...",
      "thumb_url": "https://cdn/...",
      "created_at": "2025-11-16T12:34:56Z"
    }
  ],
  "created_at": "2025-11-16T12:34:56Z",
  "updated_at": "2025-11-16T12:34:56Z"
}
```

### 3. Статусы Violation

Добавлен новый статус:
- `new` - новое нарушение
- `confirmed` - подтверждено
- `resolved` - полностью решено
- `partially_resolved` - частично решено (новый)

### 4. API эндпоинты

#### `GET /api/violations/{id}` - Получение нарушения

**Изменения:**
- Фото теперь возвращаются из всех заявок (включая заявку `open` при создании и заявки на закрытие)
- **ДОБАВЛЕНО:** Поле `requests[]` с массивом всех заявок (с датой создания, автором, фото и описанием)
- Формат ответа расширен, но сохраняет обратную совместимость:

```json
{
  "user_id": 123,
  "description": "Мусор вдоль дороги",
  "lat": 55.750733,
  "lng": 37.617761,
  "photos": [
    {
      "id": "uuid",
      "violation_id": "uuid",
      "url": "https://cdn/...",
      "thumb_url": "https://cdn/..."
    }
  ],
  "requests": [
    {
      "id": "uuid",
      "status": "open",
      "created_by_user_id": 123,
      "comment": "",
      "created_at": "2025-11-25T21:06:28Z",
      "photos": [
        {
          "id": "uuid",
          "violation_id": "uuid",
          "url": "https://cdn/...",
          "thumb_url": "https://cdn/..."
        }
      ]
    },
    {
      "id": "uuid",
      "status": "closed",
      "created_by_user_id": 456,
      "comment": "Проблема решена",
      "created_at": "2025-11-26T10:30:00Z",
      "photos": [
        {
          "id": "uuid",
          "violation_id": "uuid",
          "url": "https://cdn/...",
          "thumb_url": "https://cdn/..."
        }
      ]
    }
  ]
}
```

**Структура заявки в массиве `requests[]`:**
- `id` (string, UUID) - идентификатор заявки
- `status` (string) - статус: `"open"`, `"partially_closed"`, `"closed"`
- `created_by_user_id` (integer) - ID пользователя, создавшего заявку
- `comment` (string, optional) - комментарий/описание заявки
- `created_at` (string, ISO 8601) - дата и время создания заявки
- `photos[]` (array) - массив фотографий этой заявки с публичными URL

**Важно:** 
- Поле `photos[]` (на верхнем уровне) содержит все фото из всех заявок для обратной совместимости
- Поле `requests[]` содержит детальную информацию о каждой заявке с её фото
- Все URL в `photos[]` внутри заявок - это публичные URL с временным доступом (24 часа)

#### `POST /api/violations` - Создание нарушения

**Без изменений:**
- Формат запроса остался прежним
- Фото загружаются через `photos[]` в multipart/form-data
- На сервере фото автоматически привязываются к заявке `open`

**Формат запроса:**
```
POST /api/violations
Content-Type: multipart/form-data

type: garbage
description: Мусор вдоль дороги
lat: 55.750733
lng: 37.617761
photos[]: file1.jpg
photos[]: file2.jpg
```

#### `POST /api/violations/{id}/close-request` - Закрытие нарушения (НОВЫЙ)

**Назначение:** Создать заявку на закрытие нарушения (частичное или полное)

**Метод:** `POST`

**URL:** `/api/violations/{id}/close-request`

**Авторизация:** Требуется (Bearer token)

**Content-Type:** `multipart/form-data`

**Параметры формы:**
- `status` (required): `partially_closed` или `closed`
- `comment` (optional): текстовый комментарий
- `photos[]` (optional): массив файлов изображений

**Пример запроса:**
```
POST /api/violations/550e8400-e29b-41d4-a716-446655440000/close-request
Content-Type: multipart/form-data

status: closed
comment: Проблема решена, мусор убран
photos[]: after1.jpg
photos[]: after2.jpg
```

**Ответ 200:**
```json
{
  "id": "uuid",
  "violation_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "closed",
  "created_by_user_id": 123,
  "comment": "Проблема решена, мусор убран",
  "photos": [
    {
      "id": "uuid",
      "request_id": "uuid",
      "url": "https://cdn/...",
      "thumb_url": "https://cdn/...",
      "created_at": "2025-11-16T12:34:56Z"
    }
  ],
  "created_at": "2025-11-16T12:34:56Z",
  "updated_at": "2025-11-16T12:34:56Z"
}
```

**Логика:**
- При создании заявки со статусом `partially_closed`: статус `Violation` автоматически меняется на `partially_resolved`
- При создании заявки со статусом `closed`: статус `Violation` автоматически меняется на `resolved`
- Можно создать несколько заявок на закрытие одного нарушения (история всех попыток)

**Ошибки:**
- `400 Bad Request` - неверный формат ID или статуса
- `401 Unauthorized` - не авторизован
- `404 Not Found` - нарушение не найдено

## Что нужно обновить в мобильном приложении

### 1. Создание нарушения
✅ **Без изменений** - формат запроса остался прежним

### 2. Получение нарушения

**Изменения:**
- ✅ Поле `photos[]` осталось для обратной совместимости (все фото из всех заявок)
- 🆕 **ДОБАВЛЕНО:** Поле `requests[]` с массивом заявок

**Рекомендация:** Используйте поле `requests[]` для отображения истории заявок:
- Заявка `open` - фото при создании нарушения
- Заявки `partially_closed` - частичные закрытия
- Заявки `closed` - полные закрытия

**Пример использования:**

```javascript
// Получение нарушения с заявками
const getViolation = async (violationId) => {
  const response = await fetch(
    `${API_BASE_URL}/api/violations/${violationId}`,
    {
      method: 'GET',
    }
  );

  const data = await response.json();
  
  // Все фото (для обратной совместимости)
  const allPhotos = data.photos;
  
  // Заявки с детальной информацией
  const requests = data.requests || [];
  
  // Найти заявку на открытие (с фото при создании)
  const openRequest = requests.find(req => req.status === 'open');
  
  // Найти заявки на закрытие
  const closedRequests = requests.filter(
    req => req.status === 'closed' || req.status === 'partially_closed'
  );
  
  return {
    violation: data,
    openRequest,
    closedRequests,
    allPhotos,
  };
};
```

### 3. Закрытие нарушения (НОВОЕ)

**Раньше:**
- `POST /violations/{id}/resolve`
- `POST /violations/{id}/partial_resolve`

**Теперь:**
- `POST /api/violations/{id}/close-request` с параметром `status`

**Пример кода (React Native):**

```javascript
// Закрытие нарушения (полное)
const closeViolation = async (violationId, comment, photos) => {
  const formData = new FormData();
  formData.append('status', 'closed');
  formData.append('comment', comment);
  
  photos.forEach((photo, index) => {
    formData.append('photos[]', {
      uri: photo.uri,
      type: 'image/jpeg',
      name: `photo_${index}.jpg`,
    });
  });

  const response = await fetch(
    `${API_BASE_URL}/api/violations/${violationId}/close-request`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
      body: formData,
    }
  );

  return response.json();
};

// Частичное закрытие
const partiallyCloseViolation = async (violationId, comment, photos) => {
  const formData = new FormData();
  formData.append('status', 'partially_closed');
  formData.append('comment', comment);
  
  photos.forEach((photo, index) => {
    formData.append('photos[]', {
      uri: photo.uri,
      type: 'image/jpeg',
      name: `photo_${index}.jpg`,
    });
  });

  const response = await fetch(
    `${API_BASE_URL}/api/violations/${violationId}/close-request`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
      body: formData,
    }
  );

  return response.json();
};
```

### 4. Отображение истории заявок

**Теперь доступно:** Поле `requests[]` в ответе `GET /api/violations/{id}` содержит полную историю всех заявок.

**Что можно отобразить:**
- История создания: заявка `open` с датой создания и автором
- История закрытий: все заявки `partially_closed` и `closed` с комментариями и фото
- Хронология: заявки отсортированы по `created_at`

**Пример отображения:**

```javascript
// Отображение истории заявок
const renderViolationHistory = (violation) => {
  const { requests } = violation;
  
  return requests.map(request => ({
    id: request.id,
    type: request.status === 'open' ? 'Создание' : 
          request.status === 'partially_closed' ? 'Частичное закрытие' : 
          'Полное закрытие',
    author: request.created_by_user_id,
    date: new Date(request.created_at),
    comment: request.comment,
    photos: request.photos,
  }));
};
```

## Миграция данных

- Существующие нарушения будут работать корректно
- При следующем создании нарушения автоматически создастся заявка `open`
- Старые фото из `violation_photos` можно мигрировать в заявки (если нужно)

## Вопросы?

Если что-то непонятно или нужны дополнительные детали - пишите!

