# QabulGo Backend — Bemorlar uchun Autentifikatsiya va Profil Tizimi

QabulGo — O‘zbekistondagi xususiy klinikalar, shifokorlar va bemorlarni bog‘lovchi yagona onlayn platforma. Ushbu modul loyihaning **1-bosqichi** bo‘lib, bemorlar uchun xavfsiz ro‘yxatdan o‘tish (Register), emailni 6 xonali OTP kod orqali tasdiqlash, tizimga kirish (Login), JWT Access va Refresh Token aylanishi (Token Rotation), parolni tiklash (Forgot/Reset Password) hamda shaxsiy profilni boshqarish funksiyalarini taqdim etadi.

---

## 🛠 Texnologiyalar Steki

- **Dasturlash tili:** Go (Golang) 1.22+
- **HTTP Framework:** Gin Gonic
- **Ma’lumotlar bazasi:** PostgreSQL 16 (UUID primary keylar, pgcrypto)
- **ORM:** GORM (Connection pooling, parameterized queries)
- **Migratsiyalar:** SQL migration fayllari (`migrations/`)
- **Autentifikatsiya:** JWT (Access Token, 15m) + Kriptografik tasodifiy Refresh Token (7 kun, SHA-256 xeshi bazada, HttpOnly Secure Cookie)
- **Parol xeshlash:** Argon2id (eng zamonaviy xavfsizlik standarti)
- **API Hujjatlari:** Swagger / OpenAPI 2.0 / 3.0 (`/swagger/index.html`)
- **Xavfsizlik:** CORS, Rate Limiting (Token Bucket), Panic Recovery, Request ID (`X-Request-ID`), Timing-attack himoyasi

---

## 📁 Loyiha Strukturasi (`back/` papkasi ichida)

```text
back/
├── cmd/
│   └── api/
│       └── main.go               # Ilova boshlang'ich nuqtasi, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go             # Environment konfiguratsiyasi
│   ├── database/
│   │   ├── database.go           # GORM ulanish va connection pool sozlamalari
│   │   └── migrate.go            # SQL migratsiyalarni avtomatik bajaruvchi modul
│   ├── models/
│   │   ├── user.go               # User modeli va rollar (patient, doctor, va h.k.)
│   │   ├── refresh_session.go    # Refresh sessiyalari modeli
│   │   └── verification_token.go # Email va reset tokenlari modeli
│   ├── dto/
│   │   ├── auth.go               # Register, Login, Verify DTOlari
│   │   ├── patient.go            # Bemor profili DTOlari
│   │   └── response.go           # Standartlashtirilgan Success/Error JSON formati
│   ├── handlers/
│   │   ├── auth_handler.go       # Auth HTTP handlerlari va cookie boshqaruvi
│   │   ├── patient_handler.go    # Bemor profili handlerlari
│   │   └── health_handler.go     # Server salomatligi (/health)
│   ├── services/
│   │   ├── auth_service.go       # Biznes mantiq (register, login, verify, reset)
│   │   ├── email_service.go      # Email xizmati (SMTP va Mock/Dev rejim)
│   │   ├── token_service.go      # JWT va Refresh token rotation
│   │   └── patient_service.go    # Profil ko'rish va tahrirlash
│   ├── repositories/
│   │   ├── user_repository.go
│   │   ├── refresh_session_repository.go
│   │   └── verification_token_repository.go
│   ├── middleware/
│   │   ├── auth.go               # Bearer JWT va Role tekshiruvchi middleware
│   │   ├── cors.go               # Cross-Origin Resource Sharing
│   │   ├── logger.go             # Maxfiy ma'lumotlarni yashiruvchi xavfsiz logger
│   │   ├── rate_limit.go         # IP bo'yicha token bucket rate limiter
│   │   ├── recovery.go           # Panic tutuvchi middleware
│   │   └── request_id.go         # X-Request-ID unikal identifikatori
│   ├── routes/
│   │   └── routes.go             # Yo'llar va middlewarelarni bog'lash
│   └── utils/
│       ├── argon2.go             # Argon2id xeshlash va tekshirish
│       ├── jwt.go                # JWT yaratish va imzo tekshirish
│       ├── random.go             # Kriptografik 6 xonali OTP va random tokenlar
│       └── validator.go          # O‘zbekiston telefoni (+998), email va parol tekshiruvi
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_refresh_sessions_table.up.sql
│   ├── 000002_create_refresh_sessions_table.down.sql
│   ├── 000003_create_verification_tokens_table.up.sql
│   └── 000003_create_verification_tokens_table.down.sql
├── docs/
│   └── docs.go                   # Swagger spetsifikatsiyasi
├── tests/
│   ├── mock_repository.go        # Unit va integratsion testlar uchun xotiradagi modellar
│   ├── auth_test.go              # 18 ta majburiy talab bo'yicha testlar
│   └── patient_test.go           # Bemor profili testlari
├── .env.example                  # Muhit sozlamalari namunasi
├── .gitignore
├── docker-compose.yml            # PostgreSQL konteyneri
├── Makefile                      # Buyruqlar avtomatizatsiyasi
├── go.mod
└── README.md
```

---

## 🚀 Ishga Tushirish Bo‘yicha Qo‘llanma (Windows PowerShell)

### 1. Go (Golang) O‘rnatish

Agar kompyuteringizda Go o‘rnatilmagan bo‘lsa:
1. [golang.org/dl](https://go.dev/dl/) saytidan Windows uchun Go o‘rnatuvchisini (`.msi`) yuklab oling va o‘rnating.
2. PowerShell'ni qayta ochib tekshiring:
   ```powershell
   go version
   ```

### 2. PostgreSQL'ni Ishga Tushirish

Docker orqali ishga tushirish (tavsiya etiladi):
```powershell
# back papkasiga o'ting
cd d:\QabulGo\back

# PostgreSQL konteynerini fonda ishga tushirish
docker compose up -d
```

*Agar kompyuteringizda Docker bo‘lmasa:* Lokal PostgreSQL xizmatini o‘rnating va `qabulgo` nomli ma’lumotlar bazasini yarating.

### 3. `.env` Faylini Sozlash

`.env.example` faylidan `.env` nusxasini yarating:
```powershell
Copy-Item .env.example .env
```

`.env` ichidagi muhim parametrlar:
- `DATABASE_URL`: PostgreSQL ulanish manzili.
- `JWT_SECRET`: Eng kamida 32 belgidan iborat tasodifiy xavfsiz kalit.
- `EMAIL_MOCK=true`: Development rejimida SMTP server sozlanmagan bo‘lsa, 6 xonali tasdiqlash kodlari to‘g‘ridan-to‘g‘ri server konsolida/terminalida ko‘rinadi.

### 4. Dependency'larni Yuklab Olish

```powershell
go mod tidy
go mod download
```

### 5. Backend Serverni Ishga Tushirish

```powershell
go run cmd/api/main.go
```

Server ishga tushganda quyidagi xabar chiqadi:
```text
QabulGo Backend server http://localhost:8080 da ishga tushdi
Swagger API hujjatlari: http://localhost:8080/swagger/index.html
```

---

## 🧪 Testlarni Ishga Tushirish

Loyihada barcha 18 ta talab etilgan biznes-ssenariylar to‘liq qamrab olingan. Testlar in-memory mock orqali PostgreSQL'siz ham to‘liq va mustaqil ishlaydi:

```powershell
go test -v ./tests/...
```

---

## 📖 Swagger API Dokumentatsiyasi

Server ishga tushgandan so‘ng brauzer orqali quyidagi manzilga kiring:
👉 **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

U yerda barcha endpointlar, so‘rov va javob namunalari, status kodlari interaktiv holda mavjud.

---

## 📡 API Endpointlar Ro‘yxati va PowerShell Orqali Sinash

Barcha muvaffaqiyatli javoblar:
```json
{
  "success": true,
  "message": "Amal muvaffaqiyatli bajarildi",
  "data": {}
}
```

Barcha xatolik javoblari:
```json
{
  "success": false,
  "message": "Xatolik tavsifi",
  "errors": {
    "field": "Aniqlangan xatolik"
  }
}
```

### 1. Bemor Ro‘yxatdan O‘tishi (Register)
- **Endpoint:** `POST /api/v1/auth/patient/register`
- **Tavsif:** Bemor ro‘yxatdan o‘tadi, avtomatik `patient` roli beriladi va emailga 6 xonali kod yuboriladi.

```powershell
$body = @{
    first_name       = "Ali"
    last_name        = "Valiyev"
    email            = "ali@example.com"
    phone            = "+998901234567"
    password         = "StrongPass123!"
    confirm_password = "StrongPass123!"
    accept_terms     = $true
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/patient/register" -Method Post -Body $body -ContentType "application/json"
```

*(Terminal logida `[DEV EMAIL MOCK] Code: XXXXXX` ko‘rinadi)*

---

### 2. Emailni Tasdiqlash (Verify Email)
- **Endpoint:** `POST /api/v1/auth/verify-email`

```powershell
$body = @{
    email = "ali@example.com"
    code  = "123456" # Terminalda chiqqan 6 xonali kod
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/verify-email" -Method Post -Body $body -ContentType "application/json"
```

---

### 3. Tizimga Kirish (Login)
- **Endpoint:** `POST /api/v1/auth/patient/login`
- **Tavsif:** Access token (JSON) va `refresh_token` (HttpOnly Cookie) qaytaradi.

```powershell
$body = @{
    email    = "ali@example.com"
    password = "StrongPass123!"
} | ConvertTo-Json

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/patient/login" -Method Post -Body $body -ContentType "application/json" -WebSession $session

$token = $loginResponse.data.access_token
Write-Host "Olingan Access Token: $token"
```

---

### 4. Bemor Profilini Ko‘rish (Protected: /patient/me)
- **Endpoint:** `GET /api/v1/patient/me`
- **Header:** `Authorization: Bearer <access_token>`

```powershell
$headers = @{
    Authorization = "Bearer $token"
}

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/patient/me" -Method Get -Headers $headers
```

---

### 5. Bemor Profilini Tahrirlash
- **Endpoint:** `PATCH /api/v1/patient/me`

```powershell
$updateBody = @{
    first_name = "Alisher"
    last_name  = "Valiyev"
    phone      = "+998909876543"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/patient/me" -Method Patch -Headers $headers -Body $updateBody -ContentType "application/json"
```

---

### 6. Access Tokenni Yangilash (Refresh)
- **Endpoint:** `POST /api/v1/auth/refresh`
- Token rotation mexanizmi: eski refresh token bekor qilinadi, yangi juftlik beriladi. Agar bekor qilingan token qayta ishlatilsa, foydalanuvchining barcha sessiyalari xavfsizlik maqsadida to‘xtatiladi.

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/refresh" -Method Post -WebSession $session
```

---

### 7. Kabinetdan Chiqish (Logout)
- **Endpoint:** `POST /api/v1/auth/logout`

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/logout" -Method Post -WebSession $session
```

---

### 8. Parolni Tiklash (Forgot / Reset Password)
- `POST /api/v1/auth/forgot-password` (email yuboriladi, 6 xonali tiklash kodi jo'natiladi)
- `POST /api/v1/auth/reset-password` (kod va yangi parol orqali parol yangilanadi va barcha eski sessiyalar bekor qilinadi)

---

## 🤝 React + TypeScript Frontend Uchun API Shartnomasi

Kelajakda `front/` papkasida yaratiladigan React frontend uchun tavsiyalar:
1. **Axios yoki Fetch sozlamasi:**
   Har bir so‘rovda `withCredentials: true` (yoki `credentials: 'include'`) parametrini yoqing. Bu brauzerga `refresh_token` HttpOnly cookie'sini avtomatik jo‘natish imkonini beradi.
2. **Access Token saqlash:**
   Access tokenni `localStorage` ga saqlamang (XSS xavfi sababli). Uni React xotirasida (masalan, Zustand yoki React Context / in-memory variable) saqlang. Sahifa yangilanganda (F5) yoki 401 Unauthorized qaytganda, frontend avtomatik `/api/v1/auth/refresh` chaqirib yangi access token oladi.
3. **CORS:**
   Backend `http://localhost:5173` (Vite) va `http://localhost:3000` ga to‘liq sozlangan.
