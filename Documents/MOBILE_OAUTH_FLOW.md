# Инструкция по OAuth Flow для мобильного приложения

## 📋 Содержание

1. [Обзор изменений](#обзор-изменений)
2. [Новый Flow](#новый-flow)
3. [API Endpoints](#api-endpoints)
4. [Обработка Deep Link](#обработка-deep-link)
5. [Примеры реализации](#примеры-реализации)
6. [PKCE (Proof Key for Code Exchange)](#pkce-proof-key-for-code-exchange)
7. [Важные замечания](#важные-замечания)
8. [Миграция со старого flow](#миграция-со-старого-flow)
9. [Тестирование](#тестирование)
10. [Troubleshooting](#troubleshooting)

## Обзор изменений

OAuth flow был изменен: теперь все провайдеры (Yandex, Google и др.) перенаправляют на серверный endpoint вместо прямого deep link в мобильное приложение. Сервер обрабатывает callback, обменивает код на токены и перенаправляет на мобильное приложение с готовыми токенами.

**Ключевые изменения:**
- ✅ Провайдеры перенаправляют на сервер (`https://api.green-warden.ru/api/auth/callback`)
- ✅ Сервер автоматически обменивает код на токены
- ✅ Мобильное приложение получает готовые токены через deep link
- ✅ `code_verifier` теперь обязателен в запросе `/api/user/login`
- ❌ Endpoint `/api/user/exchange` больше не нужен для основного login flow

## Новый Flow

```
1. Mobile App → POST /api/user/login (code_challenge, code_verifier, provider)
   ↓
2. Server → возвращает auth_url и state
   ↓
3. Mobile App → открывает auth_url в браузере
   ↓
4. Provider → перенаправляет на https://server.com/api/auth/callback?code=...&state=...&provider=...
   ↓
5. Server → обменивает code на токены, сохраняет сессию
   ↓
6. Server → перенаправляет на warden://auth/callback?provider=...&access_token=...&user_id=...
   ↓
7. Mobile App → получает токены через deep link
```

## Что изменилось

### Старый flow (больше не используется):
- Провайдер перенаправлял напрямую на `warden://auth/callback?code=...&state=...`
- Мобильное приложение получало `code` и вызывало `POST /api/user/exchange` с `code` и `code_verifier`

### Новый flow:
- Провайдер перенаправляет на серверный endpoint
- Сервер обменивает код на токены автоматически
- Мобильное приложение получает готовые токены через deep link
- **Endpoint `/api/user/exchange` больше не нужен для основного login flow**

## API Endpoints

### 1. POST /api/user/login

**Запрос:**
```json
{
  "provider": "yandex",  // или "google"
  "code_challenge": "BASE64URL(SHA256(code_verifier))",
  "code_verifier": "случайная_строка_43-128_символов"
}
```

**Ответ:**
```json
{
  "auth_url": "https://oauth.yandex.com/authorize?client_id=...&code_challenge=...&state=...",
  "state": "случайный_state_токен"
}
```

**Важно:**
- `code_verifier` теперь **обязателен** в запросе
- `code_verifier` должен быть сохранен локально (не отправляется на сервер после этого)
- `state` возвращается сервером и используется для валидации

### 2. GET /api/auth/callback (обрабатывается сервером автоматически)

Этот endpoint обрабатывается сервером. Мобильное приложение **не вызывает его напрямую**.

Сервер перенаправляет на deep link:
```
warden://auth/callback?provider={provider}&access_token={token}&user_id={id}
```

Или в случае ошибки:
```
warden://auth/callback?provider={provider}&error={error}&error_description={description}
```

## Обработка Deep Link

### Успешный ответ

Deep link содержит:
- `provider` - название провайдера (yandex, google)
- `access_token` - access token для API запросов
- `user_id` - ID пользователя

**Пример:**
```
warden://auth/callback?provider=yandex&access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...&user_id=123
```

**Обработка:**
1. Извлечь `access_token` и `user_id` из query параметров
2. Сохранить `access_token` для последующих API запросов
3. Refresh token сохраняется в HTTP-only cookie на сервере (автоматически)

### Ошибки

Deep link содержит:
- `provider` - название провайдера
- `error` - код ошибки
- `error_description` - описание ошибки (опционально)

**Возможные ошибки:**
- `invalid_request` - отсутствуют обязательные параметры (code, state)
- `invalid_state` - state не найден на сервере
- `expired_state` - state истек (TTL 15 минут)
- `invalid_provider` - несоответствие provider
- `exchange_failed` - ошибка при обмене кода на токены

**Пример ошибки:**
```
warden://auth/callback?provider=yandex&error=expired_state&error_description=state_expired
```

## Примеры реализации

### Android (Kotlin)

#### Шаг 1: Инициализация OAuth

```kotlin
// Data classes
data class LoginRequest(
    val provider: String,
    val code_challenge: String,
    val code_verifier: String
)

data class LoginResponse(
    val auth_url: String,
    val state: String
)

// OAuth Manager
class OAuthManager(private val apiService: ApiService) {
    
    fun startOAuthLogin(provider: String, context: Context) {
        // 1. Генерируем code_verifier и code_challenge
        val codeVerifier = generateCodeVerifier()
        val codeChallenge = generateCodeChallenge(codeVerifier)
        
        // 2. Отправляем запрос на сервер
        val request = LoginRequest(
            provider = provider,
            code_challenge = codeChallenge,
            code_verifier = codeVerifier
        )
        
        apiService.login(request).enqueue(object : Callback<LoginResponse> {
            override fun onResponse(
                call: Call<LoginResponse>, 
                response: Response<LoginResponse>
            ) {
                if (response.isSuccessful) {
                    val authUrl = response.body()?.auth_url
                    authUrl?.let {
                        // 3. Открываем auth_url в браузере
                        openBrowser(context, it)
                    }
                } else {
                    // Обработка ошибки HTTP
                    handleError("Login request failed: ${response.code()}")
                }
            }
            
            override fun onFailure(call: Call<LoginResponse>, t: Throwable) {
                handleError("Network error: ${t.message}")
            }
        })
    }
    
    private fun openBrowser(context: Context, url: String) {
        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
        context.startActivity(intent)
    }
}
```

#### Шаг 2: Обработка Deep Link

```kotlin
// В MainActivity или специальной OAuthActivity
class MainActivity : AppCompatActivity() {
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        handleDeepLink(intent)
    }
    
    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handleDeepLink(intent)
    }
    
    private fun handleDeepLink(intent: Intent) {
        val data = intent.data
        data?.let { uri ->
            if (uri.scheme == "warden" && 
                uri.host == "auth" && 
                uri.path == "/callback") {
                
                val provider = uri.getQueryParameter("provider")
                val accessToken = uri.getQueryParameter("access_token")
                val userId = uri.getQueryParameter("user_id")?.toIntOrNull()
                val error = uri.getQueryParameter("error")
                val errorDescription = uri.getQueryParameter("error_description")
                
                if (error != null) {
                    // Обработка ошибки
                    handleOAuthError(provider, error, errorDescription)
                } else if (accessToken != null && userId != null) {
                    // Успешная авторизация
                    TokenManager.saveAccessToken(accessToken)
                    UserManager.saveUserId(userId)
                    navigateToMainScreen()
                } else {
                    handleOAuthError(provider, "invalid_response", "Missing tokens")
                }
            }
        }
    }
    
    private fun handleOAuthError(
        provider: String?, 
        error: String, 
        description: String?
    ) {
        // Показать ошибку пользователю
        Toast.makeText(
            this, 
            "OAuth error: $error - ${description ?: ""}", 
            Toast.LENGTH_LONG
        ).show()
    }
}
```

#### AndroidManifest.xml

```xml
<activity
    android:name=".MainActivity"
    android:exported="true">
    <intent-filter>
        <action android:name="android.intent.action.VIEW" />
        <category android:name="android.intent.category.DEFAULT" />
        <category android:name="android.intent.category.BROWSABLE" />
        <data
            android:scheme="warden"
            android:host="auth"
            android:path="/callback" />
    </intent-filter>
</activity>
```

### iOS (Swift)

#### Шаг 1: Инициализация OAuth

```swift
import UIKit
import CryptoKit

struct LoginRequest: Codable {
    let provider: String
    let code_challenge: String
    let code_verifier: String
}

struct LoginResponse: Codable {
    let auth_url: String
    let state: String
}

class OAuthManager {
    private let apiService: ApiService
    
    init(apiService: ApiService) {
        self.apiService = apiService
    }
    
    func startOAuthLogin(provider: String) {
        // 1. Генерируем code_verifier и code_challenge
        let codeVerifier = generateCodeVerifier()
        let codeChallenge = generateCodeChallenge(verifier: codeVerifier)
        
        // 2. Отправляем запрос на сервер
        let request = LoginRequest(
            provider: provider,
            code_challenge: codeChallenge,
            code_verifier: codeVerifier
        )
        
        apiService.login(request: request) { [weak self] result in
            switch result {
            case .success(let response):
                // 3. Открываем auth_url в браузере
                if let url = URL(string: response.auth_url) {
                    UIApplication.shared.open(url)
                }
            case .failure(let error):
                print("Login error: \(error)")
            }
        }
    }
}
```

#### Шаг 2: Обработка Deep Link

```swift
// В SceneDelegate или AppDelegate
import UIKit

class SceneDelegate: UIResponder, UIWindowSceneDelegate {
    
    func scene(_ scene: UIScene, openURLContexts URLContexts: Set<UIOpenURLContext>) {
        guard let url = URLContexts.first?.url else { return }
        handleDeepLink(url: url)
    }
    
    private func handleDeepLink(url: URL) {
        guard url.scheme == "warden",
              url.host == "auth",
              url.path == "/callback" else {
            return
        }
        
        let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
        let queryItems = components?.queryItems ?? []
        
        let provider = queryItems.first(where: { $0.name == "provider" })?.value
        let accessToken = queryItems.first(where: { $0.name == "access_token" })?.value
        let userIdString = queryItems.first(where: { $0.name == "user_id" })?.value
        let error = queryItems.first(where: { $0.name == "error" })?.value
        let errorDescription = queryItems.first(where: { $0.name == "error_description" })?.value
        
        if let error = error {
            // Обработка ошибки
            handleOAuthError(provider: provider, error: error, description: errorDescription)
        } else if let accessToken = accessToken,
                  let userIdString = userIdString,
                  let userId = Int(userIdString) {
            // Успешная авторизация
            TokenManager.shared.saveAccessToken(accessToken)
            UserManager.shared.saveUserId(userId)
            navigateToMainScreen()
        } else {
            handleOAuthError(provider: provider, error: "invalid_response", description: "Missing tokens")
        }
    }
    
    private func handleOAuthError(provider: String?, error: String, description: String?) {
        // Показать ошибку пользователю
        print("OAuth error: \(error) - \(description ?? "")")
    }
}
```

#### Info.plist

```xml
<key>CFBundleURLTypes</key>
<array>
    <dict>
        <key>CFBundleTypeRole</key>
        <string>Editor</string>
        <key>CFBundleURLName</key>
        <string>com.yourapp.warden</string>
        <key>CFBundleURLSchemes</key>
        <array>
            <string>warden</string>
        </array>
    </dict>
</array>
```

## PKCE (Proof Key for Code Exchange)

### Android (Kotlin)

#### Генерация code_verifier

```kotlin
import java.security.SecureRandom
import android.util.Base64

fun generateCodeVerifier(): String {
    val bytes = ByteArray(32) // 32 байта = 43 символа в base64url
    SecureRandom().nextBytes(bytes)
    return Base64.encodeToString(bytes, Base64.URL_SAFE or Base64.NO_WRAP)
        .substring(0, 43) // Обрезаем до 43 символов (RFC 7636 требует 43-128)
        .replace("=", "") // Убираем padding
}
```

#### Генерация code_challenge

```kotlin
import java.security.MessageDigest
import android.util.Base64

fun generateCodeChallenge(verifier: String): String {
    val digest = MessageDigest.getInstance("SHA-256")
    val hash = digest.digest(verifier.toByteArray(Charsets.UTF_8))
    return Base64.encodeToString(hash, Base64.URL_SAFE or Base64.NO_WRAP)
        .replace("=", "") // Убираем padding
}
```

### iOS (Swift)

#### Генерация code_verifier

```swift
import Foundation
import CryptoKit

func generateCodeVerifier() -> String {
    var bytes = [UInt8](repeating: 0, count: 32)
    _ = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
    
    let data = Data(bytes)
    let base64 = data.base64EncodedString()
        .replacingOccurrences(of: "+", with: "-")
        .replacingOccurrences(of: "/", with: "_")
        .replacingOccurrences(of: "=", with: "")
    
    // RFC 7636 требует 43-128 символов, берем 43
    return String(base64.prefix(43))
}
```

#### Генерация code_challenge

```swift
import Foundation
import CryptoKit

func generateCodeChallenge(verifier: String) -> String {
    guard let data = verifier.data(using: .utf8) else {
        fatalError("Failed to convert verifier to data")
    }
    
    let hash = SHA256.hash(data: data)
    let base64 = Data(hash).base64EncodedString()
        .replacingOccurrences(of: "+", with: "-")
        .replacingOccurrences(of: "/", with: "_")
        .replacingOccurrences(of: "=", with: "")
    
    return base64
}
```

## Важные замечания

1. **code_verifier обязателен**: Теперь `code_verifier` должен передаваться в `/api/user/login` вместе с `code_challenge`

2. **State валидация**: Хотя сервер валидирует state автоматически, можно дополнительно проверить его на клиенте (опционально)

3. **Токены**: 
   - `access_token` приходит в deep link и должен быть сохранен
   - `refresh_token` сохраняется в HTTP-only cookie на сервере
   - Для refresh используйте `POST /api/user/refresh` (как и раньше)

4. **Безопасность**:
   - `code_verifier` должен быть случайным и уникальным для каждого запроса
   - Не логируйте `code_verifier` или `access_token`
   - Используйте secure storage для сохранения токенов

5. **Обработка ошибок**: Всегда проверяйте наличие `error` параметра в deep link перед обработкой успешного ответа

## Миграция со старого flow

Если у вас был старый код, который использовал `/api/user/exchange`:

1. **Удалите** вызовы `POST /api/user/exchange` после получения кода из deep link
2. **Добавьте** `code_verifier` в запрос `POST /api/user/login`
3. **Измените** обработку deep link: теперь получаете `access_token` напрямую, а не `code`
4. **Удалите** логику обмена кода на токены на клиенте

## Тестирование

Для тестирования убедитесь, что:
- `API_ROOT` установлен в переменных окружения сервера (например, `https://api.green-warden.ru`)
  - Это адрес API сервера, куда провайдеры будут перенаправлять после авторизации
  - Если не установлен, используется fallback: `https://api.green-warden.ru`
- В настройках OAuth провайдеров указан правильный redirect URI:
  - Yandex: `https://api.green-warden.ru/api/auth/callback?provider=yandex`
  - Google: `https://api.green-warden.ru/api/auth/callback?provider=google`
- Deep link `warden://auth/callback` зарегистрирован в манифесте приложения

**Важно:** `API_ROOT` должен указывать на адрес API сервера, а не на клиентский домен. Например:
- Клиент: `https://green-warden.ru`
- API сервер: `https://api.green-warden.ru` (используется для OAuth redirect)

## Troubleshooting

### Проблема: Deep link не открывается в приложении

**Решение:**
- Проверьте регистрацию deep link в манифесте (Android) или Info.plist (iOS)
- Убедитесь, что схема `warden://` правильно зарегистрирована
- Проверьте, что приложение установлено и deep link зарегистрирован в системе

### Проблема: Ошибка "invalid_state"

**Причины:**
- State истек (TTL 15 минут)
- State был использован дважды
- Несоответствие state между запросом и callback

**Решение:**
- Убедитесь, что пользователь завершает авторизацию в течение 15 минут
- Не используйте один state дважды
- Проверьте, что state правильно передается через весь flow

### Проблема: Ошибка "exchange_failed"

**Причины:**
- Неверный `code_verifier`
- Код уже использован
- Проблемы с настройками OAuth провайдера

**Решение:**
- Проверьте правильность генерации `code_verifier` и `code_challenge`
- Убедитесь, что `code_verifier` передается в `/api/user/login`
- Проверьте логи сервера для детальной информации об ошибке
- Проверьте настройки OAuth провайдера (client_id, client_secret, redirect_uri)

### Проблема: Ошибка "invalid_provider"

**Причины:**
- Несоответствие provider между запросом и callback
- Provider не поддерживается

**Решение:**
- Убедитесь, что используете правильное название провайдера ("yandex", "google")
- Проверьте, что provider одинаковый в запросе `/api/user/login` и в redirect URI

### Проблема: Токены не сохраняются

**Решение:**
- Проверьте обработку deep link - убедитесь, что извлекаете `access_token` и `user_id`
- Используйте secure storage для сохранения токенов
- Проверьте, что refresh token сохраняется в cookie (это делается автоматически на сервере)

### Проблема: Redirect URI не совпадает

**Решение:**
- Убедитесь, что в настройках OAuth провайдера указан правильный redirect URI:
  - Yandex: `https://api.green-warden.ru/api/auth/callback?provider=yandex`
  - Google: `https://api.green-warden.ru/api/auth/callback?provider=google`
- Проверьте, что `API_ROOT` правильно установлен в переменных окружения сервера

## Чек-лист для разработчика

Перед релизом убедитесь, что:

- [ ] `code_verifier` генерируется правильно (43-128 символов, base64url)
- [ ] `code_challenge` вычисляется как SHA256(code_verifier) в base64url
- [ ] `code_verifier` передается в запрос `/api/user/login`
- [ ] Deep link `warden://auth/callback` зарегистрирован в манифесте/Info.plist
- [ ] Обработка deep link корректно извлекает `access_token` и `user_id`
- [ ] Обработка ошибок реализована для всех возможных error кодов
- [ ] Токены сохраняются в secure storage
- [ ] Логирование не содержит чувствительных данных (code_verifier, access_token)
- [ ] Тестирование проведено для всех поддерживаемых провайдеров

## Поддержка

При возникновении проблем проверьте:
1. Правильность генерации `code_verifier` и `code_challenge`
2. Корректность обработки deep link
3. Логи сервера для диагностики ошибок обмена токенов
4. Настройки OAuth провайдеров (redirect URI должен совпадать)
5. Переменные окружения сервера (`API_ROOT` должен быть установлен)

**Контакты для поддержки:**
- Техническая документация: см. код сервера
- Логи сервера: проверьте логи приложения на сервере