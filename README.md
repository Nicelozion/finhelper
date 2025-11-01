# FinHelper

Приложение для управления финансами.

## Структура проекта

Проект разделен на две основные части:
- **frontend/** - React + TypeScript приложение
- **backend/** - Backend API (будет разработан параллельно)

## Разработка

### Frontend

Перейдите в директорию frontend и следуйте инструкциям в `frontend/README.md`:

```bash
cd frontend
pnpm install
pnpm run dev
```

### Backend

Полная инструкция по реализации backend API для интеграции с frontend находится в `backend/README.md`.

Кратко:
- Backend должен работать на `http://localhost:8080`
- Обязательна настройка CORS для `http://localhost:5173`
- Все endpoints должны возвращать заголовок `X-Request-Id`
- Endpoints: `POST /api/banks/{bank}/connect`, `GET /api/accounts`, `GET /api/transactions`

Подробности см. в `backend/README.md`

## Команда

Проект разрабатывается командой, где разные участники работают над frontend и backend параллельно.
