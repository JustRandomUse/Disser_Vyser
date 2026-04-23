# Деплой на Vercel - Пошаговая инструкция

Этот проект использует **два отдельных Vercel проекта** из одного Git-репозитория:
1. **Backend** (Go API) - отдает только API endpoints
2. **Frontend** (Vue 3) - публичный сайт с rewrite на backend

---

## Предварительные требования

- Аккаунт на [Vercel](https://vercel.com)
- Git-репозиторий проекта подключен к Vercel
- API ключ от Sensor Hub API

---

## Шаг 1: Деплой Backend

### 1.1 Создать Backend Project в Vercel

1. Зайти на https://vercel.com/new
2. Выбрать ваш Git-репозиторий
3. Настроить проект:
   - **Project Name**: `air-quality-backend` (или любое имя)
   - **Framework Preset**: Other
   - **Root Directory**: `back` ⚠️ **ВАЖНО**
   - **Build Command**: оставить пустым (Vercel автоматически определит Go)
   - **Output Directory**: оставить пустым

### 1.2 Настроить Environment Variables

В настройках проекта (Settings → Environment Variables) добавить:

```
SENSOR_API_KEY = ваш-реальный-api-ключ
ALLOWED_ORIGINS = https://your-frontend-domain.vercel.app
```

⚠️ **ВАЖНО**: `ALLOWED_ORIGINS` нужно будет обновить после деплоя frontend (см. Шаг 2.3)

### 1.3 Задеплоить

Нажать **Deploy**. Vercel соберет Go приложение и задеплоит его.

### 1.4 Проверить Backend

После успешного деплоя:

1. Скопировать URL деплоя (например: `https://air-quality-backend.vercel.app`)
2. Проверить health check:
   ```
   curl https://air-quality-backend.vercel.app/api/health
   ```
   Должен вернуть: `{"status":"ok"}`

3. Проверить API:
   ```
   curl https://air-quality-backend.vercel.app/api/datasets
   ```

✅ Если API отвечает - backend готов!

---

## Шаг 2: Деплой Frontend

### 2.1 Обновить vercel.json

Открыть `front/vercel.json` и заменить placeholder на реальный URL backend:

```json
{
  "version": 2,
  "buildCommand": "npm run build",
  "outputDirectory": "dist",
  "rewrites": [
    {
      "source": "/api/:path*",
      "destination": "https://air-quality-backend.vercel.app/api/:path*"
    }
  ]
}
```

⚠️ Замените `https://air-quality-backend.vercel.app` на ваш реальный backend URL из Шага 1.4

Закоммитить изменения:
```bash
git add front/vercel.json
git commit -m "Configure frontend rewrite to backend"
git push
```

### 2.2 Создать Frontend Project в Vercel

1. Зайти на https://vercel.com/new
2. Выбрать тот же Git-репозиторий
3. Настроить проект:
   - **Project Name**: `air-quality-monitor` (или любое имя)
   - **Framework Preset**: Vite
   - **Root Directory**: `front` ⚠️ **ВАЖНО**
   - **Build Command**: `npm run build`
   - **Output Directory**: `dist`

### 2.3 Задеплоить Frontend

Нажать **Deploy**. Vercel соберет Vue приложение и задеплоит его.

После успешного деплоя:
1. Скопировать URL (например: `https://air-quality-monitor.vercel.app`)
2. Открыть в браузере - должен загрузиться сайт с картой

### 2.4 Обновить CORS в Backend

Вернуться в настройки **backend проекта** (Settings → Environment Variables) и обновить:

```
ALLOWED_ORIGINS = https://air-quality-monitor.vercel.app
```

⚠️ Замените на ваш реальный frontend URL

Затем **Redeploy** backend проект (Deployments → три точки → Redeploy)

---

## Шаг 3: Проверка

### 3.1 Открыть Frontend URL

Открыть `https://air-quality-monitor.vercel.app` (ваш frontend URL)

### 3.2 Проверить работу API

1. Открыть DevTools (F12) → Network
2. Обновить страницу
3. Проверить, что запросы к `/api/...` проходят успешно (статус 200)

### 3.3 Проверить карту

- Должна загрузиться карта Красноярска
- Должны отображаться датчики
- При клике на датчик должна открываться модалка с графиками

✅ Если всё работает - деплой завершен!

---

## Итоговая архитектура

```
Пользователь
    ↓
https://air-quality-monitor.vercel.app (Frontend)
    ↓ /api/* requests
https://air-quality-backend.vercel.app/api/* (Backend)
    ↓
Sensor Hub API (внешний источник данных)
```

**Важно**: 
- Пользователи видят только frontend URL
- Все API запросы проксируются через Vercel rewrites
- Backend URL не виден пользователям
- API ключ остается только на backend

---

## Обновление после изменений

### Backend изменения:
```bash
git push
# Vercel автоматически задеплоит backend
```

### Frontend изменения:
```bash
git push
# Vercel автоматически задеплоит frontend
```

---

## Troubleshooting

### Проблема: CORS ошибки

**Решение**: Проверить `ALLOWED_ORIGINS` в backend environment variables. Должен содержать точный frontend URL (с https://, без trailing slash).

### Проблема: API запросы возвращают 404

**Решение**: Проверить `front/vercel.json` - URL backend должен быть правильным.

### Проблема: Backend возвращает 500

**Решение**: Проверить логи backend в Vercel (Deployments → View Function Logs). Возможно, не установлен `SENSOR_API_KEY`.

### Проблема: Пустая карта

**Решение**: Открыть DevTools → Console. Проверить ошибки. Возможно, внешний Sensor Hub API недоступен или вернул ошибку.

---

## Custom Domain (опционально)

Если хотите использовать свой домен:

1. В настройках **frontend проекта**: Settings → Domains → Add Domain
2. Добавить ваш домен (например: `airquality.example.com`)
3. Настроить DNS записи согласно инструкциям Vercel
4. Обновить `ALLOWED_ORIGINS` в backend на новый домен
5. Redeploy backend

---

## Безопасность

✅ API ключ хранится только в backend environment variables  
✅ Frontend не имеет доступа к API ключу  
✅ CORS настроен только на ваш frontend домен  
✅ Backend не раздает статику - только API  

---

## Контакты и поддержка

При проблемах с деплоем:
1. Проверить логи в Vercel Dashboard
2. Проверить DevTools → Console в браузере
3. Проверить DevTools → Network для API запросов
