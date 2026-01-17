# Ответы на вопросы по Refresh Token API

## 1. OAuth Callback Deep Link

### Какие параметры передаются в deep link callback после успешной OAuth авторизации?

**Ответ:** После успешной OAuth авторизации в deep link передаются только:
- `provider` - название провайдера (yandex, google)
- `access_token` - JWT access token
- `user_id` - ID пользователя

**Пример успешного deep link:**
```
warden://auth/callback?provider=yandex&access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...&user_id=123
```

### Передаётся ли refresh_token как параметр URL?

**Ответ:** ❌ **НЕТ**, `refresh_token` **НЕ передается** в deep link как параметр URL.

### Или refresh_token передаётся только через HTTP-only cookie?

**Ответ:** ✅ **ДА**, `refresh_token` сохраняется **только в HTTP-only cookie** (сессия) на сервере.

**Код подтверждения:** `Server/internal/app/http/oauth_callback.go:185`
```go
session.Values[defenitions.Token] = authData.RefreshToken
session.Save(r, w)
```

---

## 2. Endpoint /api/user/refresh

### Как он работает?

**Ответ:** Endpoint работает следующим образом:
1. Извлекает `refresh_token` из **HTTP-only cookie** (сессии)
2. Валидирует refresh token
3. Генерирует новые `access_token` и `refresh_token`
4. Сохраняет новый `refresh_token` обратно в cookie
5. Возвращает оба токена в JSON ответе

### Принимает ли refresh_token в теле запроса JSON?

**Ответ:** ❌ **НЕТ**, endpoint **НЕ принимает** `refresh_token` в JSON body.

**Код подтверждения:** `Server/internal/app/http/refresh_token.go:38-48`
```go
session, err := h.store.Get(r, defenitions.SessionAuthenticationName)
if err != nil {
    http.Error(w, "Unauthorized not session", http.StatusUnauthorized)
    return
}

tokenString, ok := session.Values[defenitions.Token].(string)
if !ok {
    http.Error(w, "Unauthorized not Token", http.StatusUnauthorized)
    return
}
```

### Или ожидает refresh_token только из HTTP-only cookie?

**Ответ:** ✅ **ДА**, endpoint ожидает `refresh_token` **только из HTTP-only cookie**.

### Какой формат ответа при успешном обновлении?

**Ответ:** Формат ответа:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": 123
}
```

Где:
- `token` - это **access_token** (новый)
- `refresh_token` - новый refresh token
- `user_id` - ID пользователя

**Код подтверждения:** `Server/internal/app/http/refresh_token.go:65-75`

---

## 3. Пример успешного refresh запроса и ответа

### Запрос

**Метод:** `POST`  
**URL:** `/api/user/refresh`  
**Headers:**
```
Content-Type: application/json
Cookie: authentication=<session_cookie_with_refresh_token>
```

**Body:** ❌ **Пустое тело** (не требуется, refresh_token берется из cookie)

### Успешный ответ

**HTTP Status:** `200 OK`  
**Content-Type:** `application/json`

**Body:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJleHAiOjE3MDAwMDAwMDB9.signature",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInRva2VuX3R5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzI3NjgwMDAwfQ.signature",
  "user_id": 123
}
```

### Какие заголовки обязательны?

**Ответ:** 
- `Cookie` - обязателен (должен содержать сессию с refresh_token)
- `Content-Type: application/json` - опционально (тело пустое)

### Какой HTTP статус при успехе/неудаче?

**Успех:**
- `200 OK` - токены успешно обновлены

**Неудача:**
- `401 Unauthorized` - во всех случаях ошибки:
  - Сессия не найдена: `"Unauthorized not session"`
  - Токен отсутствует в сессии: `"Unauthorized not Token"`
  - Токен невалиден/истек: `"Unauthorized not session"`
- `500 Internal Server Error` - ошибка сохранения сессии

---

## 4. Обработка ошибок

### При какой ошибке нужно делать logout пользователя?

**Ответ:** Logout нужно делать при:
1. **401 Unauthorized** с любым текстом ошибки от `/api/user/refresh`
2. Это означает, что:
   - Сессия истекла/не найдена
   - Refresh token истек/невалиден
   - Пользователь должен авторизоваться заново

### Какой текст ошибки возвращается если refresh token истёк?

**Ответ:** При истечении refresh token возвращается:
- **HTTP Status:** `401 Unauthorized`
- **Текст ошибки:** `"Unauthorized not session"`

**Примечание:** Сервер не различает разные типы ошибок валидации токена (истек, невалиден, неправильный формат) - все возвращают одинаковый текст `"Unauthorized not session"`.

**Код подтверждения:** `Server/internal/app/http/refresh_token.go:50-54`
```go
authData, err := h.service.RefreshToken(ctx, tokenString)
if err != nil {
    http.Error(w, "Unauthorized not session", http.StatusUnauthorized)
    return
}
```

**Валидация токена:** `Server/internal/app/token_service/token_service.go:47-64`
- JWT библиотека (`jwt.ParseWithClaims`) автоматически проверяет срок действия токена
- При истечении токена возвращается ошибка, которая пробрасывается наверх

### Какой HTTP статус (401? 403? 400?)?

**Ответ:** **401 Unauthorized** - используется для всех ошибок авторизации:
- Отсутствие сессии
- Отсутствие токена в сессии
- Невалидный/истекший refresh token

---

## Дополнительная информация

### Время жизни токенов

- **Access Token:** 1 час (`1*time.Hour`)
- **Refresh Token:** 30 дней (`30*24*time.Hour`)

**Код подтверждения:** `Server/internal/app/app.go:199-200`

### Важные замечания для мобильного приложения

1. **После OAuth авторизации:**
   - Сохраняйте только `access_token` и `user_id` из deep link
   - `refresh_token` автоматически сохраняется в HTTP-only cookie браузером/WebView
   - При последующих запросах cookie автоматически отправляется

2. **При обновлении токена:**
   - Отправляйте POST запрос на `/api/user/refresh` **без тела**
   - Cookie с refresh_token должен автоматически отправляться
   - Сохраняйте новые `token` и `refresh_token` из ответа

3. **Обработка ошибок:**
   - При получении 401 от `/api/user/refresh` → выполнить logout
   - Перенаправить пользователя на экран авторизации

4. **Проблема с мобильным приложением:**
   - Если мобильное приложение не может использовать HTTP-only cookies (например, нативный клиент без WebView), то текущая реализация **не будет работать**
   - В этом случае нужно либо:
     - Изменить endpoint для принятия `refresh_token` в JSON body
     - Или использовать другой механизм передачи refresh_token
