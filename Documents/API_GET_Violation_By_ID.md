# API: Получение нарушения по ID

## Endpoint
```
GET /api/violations/{id}
```

## Описание
Публичный endpoint для получения детальной информации о нарушении по его идентификатору. Не требует авторизации.

## Параметры запроса

### Path параметры
- `id` (string, required) — UUID нарушения

**Пример:**
```
GET /api/violations/550e8400-e29b-41d4-a716-446655440000
```

## Заголовки запроса
Не требуются. Endpoint публичный.

## Успешный ответ (200 OK)

### Формат ответа
```json
{
  "user_id": 123,
  "description": "Мусор вдоль дороги",
  "lat": 55.750733,
  "lng": 37.617761,
  "photos": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "violation_id": "550e8400-e29b-41d4-a716-446655440000",
      "url": "https://cdn.example.com/violations/photo1.jpg",
      "thumb_url": "https://cdn.example.com/violations/thumb_photo1.jpg"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "violation_id": "550e8400-e29b-41d4-a716-446655440000",
      "url": "https://cdn.example.com/violations/photo2.jpg",
      "thumb_url": "https://cdn.example.com/violations/thumb_photo2.jpg"
    }
  ]
}
```

### Описание полей ответа

| Поле | Тип | Описание |
|------|-----|----------|
| `user_id` | integer | ID пользователя, создавшего нарушение |
| `description` | string | Текстовое описание нарушения (может быть пустой строкой) |
| `lat` | float | Широта места нарушения (WGS84) |
| `lng` | float | Долгота места нарушения (WGS84) |
| `photos` | array | Массив фотографий нарушения |
| `photos[].id` | string (UUID) | Идентификатор фотографии |
| `photos[].violation_id` | string (UUID) | Идентификатор нарушения |
| `photos[].url` | string | URL оригинального изображения |
| `photos[].thumb_url` | string | URL миниатюры (может быть пустой строкой) |

## Ошибки

### 400 Bad Request
Неверный формат ID нарушения.

**Пример ответа:**
```json
{
  "error": "invalid violation ID format"
}
```

### 404 Not Found
Нарушение с указанным ID не найдено.

**Пример ответа:**
```json
{
  "error": "violation not found"
}
```

## Примеры использования

### cURL
```bash
curl -X GET "https://api.example.com/api/violations/550e8400-e29b-41d4-a716-446655440000"
```

### Swift (iOS)
```swift
func getViolation(id: String) async throws -> Violation {
    let url = URL(string: "https://api.example.com/api/violations/\(id)")!
    let (data, response) = try await URLSession.shared.data(from: url)
    
    guard let httpResponse = response as? HTTPURLResponse else {
        throw APIError.invalidResponse
    }
    
    if httpResponse.statusCode == 404 {
        throw APIError.notFound
    }
    
    guard httpResponse.statusCode == 200 else {
        throw APIError.serverError
    }
    
    let decoder = JSONDecoder()
    return try decoder.decode(Violation.self, from: data)
}
```

### Kotlin (Android)
```kotlin
suspend fun getViolation(id: String): Violation {
    val url = "https://api.example.com/api/violations/$id"
    val response = httpClient.get(url)
    
    if (response.status == HttpStatusCode.NotFound) {
        throw NotFoundException("Violation not found")
    }
    
    return response.body()
}
```

### TypeScript/JavaScript
```typescript
async function getViolation(id: string): Promise<Violation> {
  const response = await fetch(`https://api.example.com/api/violations/${id}`);
  
  if (response.status === 404) {
    throw new Error('Violation not found');
  }
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  return await response.json();
}
```

## Модель данных

### Violation (ответ)
```typescript
interface Violation {
  user_id: number;
  description: string;
  lat: number;
  lng: number;
  photos: ViolationPhoto[];
}

interface ViolationPhoto {
  id: string;
  violation_id: string;
  url: string;
  thumb_url: string;
}
```

## Примечания

1. **Публичный доступ**: Endpoint не требует авторизации, доступен всем пользователям.
2. **Минимальный набор полей**: Endpoint возвращает только необходимые для отображения поля. Для получения полной информации (статус, тип, количество подтверждений и т.д.) используйте `GET /api/violations` с фильтрами.
3. **Пустой массив photos**: Если у нарушения нет фотографий, поле `photos` будет пустым массивом `[]`.
4. **Пустое описание**: Поле `description` может быть пустой строкой, если пользователь не указал описание при создании нарушения.
5. **Миниатюры**: Поле `thumb_url` может быть пустой строкой, если миниатюра не была сгенерирована.

## Связанные endpoints

- `GET /api/violations` — список нарушений с фильтрами и пагинацией
- `POST /api/violations` — создание нового нарушения (требует авторизации)


