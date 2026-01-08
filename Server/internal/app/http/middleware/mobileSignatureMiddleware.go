package middleware

import (
	"net/http"
	// Временно закомментированы импорты, так как проверка подписи отключена
	// "crypto/hmac"
	// "crypto/sha256"
	// "encoding/base64"
	// "strconv"
	// "time"
)

type MobileSignatureMiddleware struct {
	h               http.Handler
	mobileAppSecret string
}

func NewMobileSignatureMiddleware(h http.Handler, mobileAppSecret string) *MobileSignatureMiddleware {
	return &MobileSignatureMiddleware{
		h:               h,
		mobileAppSecret: mobileAppSecret,
	}
}

func (m *MobileSignatureMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Пропускаем OPTIONS (CORS preflight)
	if r.Method == "OPTIONS" {
		m.h.ServeHTTP(w, r)
		return
	}

	// 2. Проверяем подпись для всех остальных запросов
	if !m.verifyMobileSignature(r) {
		http.Error(w, "Unauthorized: Invalid mobile signature", http.StatusUnauthorized)
		return
	}

	m.h.ServeHTTP(w, r)
}

func (m *MobileSignatureMiddleware) verifyMobileSignature(r *http.Request) bool {
	// ВРЕМЕННО ОТКЛЮЧЕНО: проверка мобильной подписи отключена до синхронизации
	// алгоритма подписи между мобильным приложением и сервером.
	// TODO: Включить обратно после синхронизации алгоритма и секрета.
	return true

	// Закомментированный код проверки подписи (для восстановления позже):
	/*
	signature := r.Header.Get("X-Mobile-Signature")
	timestampStr := r.Header.Get("X-Mobile-Timestamp")

	if signature == "" || timestampStr == "" {
		return false
	}

	// Если секрет не установлен в конфиге - пропускаем проверку (для разработки)
	if m.mobileAppSecret == "" {
		return true
	}

	// Проверка timestamp (защита от replay атак)
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().UnixMilli()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	if diff > 300000 { // 5 минут (300000 мс)
		return false
	}

	// Формирование строки для подписи: только TIMESTAMP (как строка)
	signString := strconv.FormatInt(timestamp, 10)

	// Генерация ожидаемой подписи
	mac := hmac.New(sha256.New, []byte(m.mobileAppSecret))
	mac.Write([]byte(signString))
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Constant-time сравнение
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
	*/
}
