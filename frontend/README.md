# FinHelper Frontend

Современное React приложение для управления финансами с агрегацией данных из нескольких банков.

## 🚀 Технологии

- **React 19** + **TypeScript** - UI библиотека и типизация
- **Vite** - сборщик и dev-сервер
- **TailwindCSS** - утилитарный CSS фреймворк
- **shadcn/ui** - компоненты на базе Radix UI
- **Recharts** - графики и диаграммы
- **Framer Motion** - анимации
- **React Query** - управление серверным состоянием
- **Axios** - HTTP клиент
- **React Router DOM** - роутинг
- **PWA** - Progressive Web App поддержка

## 📦 Установка

```bash
pnpm install
```

## ⚙️ Настройка окружения

Создайте файл `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

Убедитесь, что `VITE_API_BASE_URL` указывает на ваш backend:

```
VITE_API_BASE_URL=http://localhost:8080
```

## 🏃 Запуск разработки

```bash
pnpm run dev
```

Приложение будет доступно по адресу `http://localhost:5173`

## 🏗️ Сборка для продакшена

```bash
pnpm run build
```

## 📱 Структура проекта

```
frontend/
├── src/
│   ├── pages/              # Страницы приложения
│   │   ├── Dashboard.tsx       # Главный экран с балансами и графиками
│   │   ├── Transactions.tsx    # Список всех транзакций с фильтрами
│   │   ├── Analytics.tsx       # Аналитика и прогнозы
│   │   ├── Settings.tsx        # Настройки и профиль
│   │   ├── Subscription.tsx     # Оформление подписки
│   │   ├── Onboarding.tsx      # Приветственный экран
│   │   └── Connect.tsx         # Подключение банков
│   ├── components/         # React компоненты
│   │   ├── ui/                 # shadcn/ui компоненты
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   └── dialog.tsx
│   │   ├── Charts/             # Компоненты графиков
│   │   │   ├── ExpensesDonut.tsx
│   │   │   └── MonthlyTrend.tsx
│   │   ├── Sidebar.tsx         # Боковое меню
│   │   ├── BankCard.tsx        # Карточка банка
│   │   ├── TransactionTable.tsx # Таблица транзакций
│   │   └── LoadingSkeleton.tsx  # Скелетоны загрузки
│   ├── hooks/              # React Query хуки
│   │   ├── useAccounts.ts
│   │   ├── useTransactions.ts
│   │   └── useConnectBank.ts
│   ├── lib/                # Утилиты и API
│   │   ├── api.ts              # API клиент с Axios и моками
│   │   ├── mockData.ts         # Моковые данные для fallback
│   │   └── utils.ts            # Утилиты (cn, и т.д.)
│   ├── types/              # TypeScript типы
│   │   └── api.ts
│   ├── App.tsx             # Главный компонент с роутингом
│   ├── main.tsx            # Точка входа
│   └── index.css           # Глобальные стили TailwindCSS
├── public/                # Публичные файлы
├── index.html             # HTML шаблон
├── tailwind.config.js     # Конфигурация TailwindCSS
├── vite.config.ts         # Конфигурация Vite + PWA
└── package.json           # Зависимости
```

## 🎨 Особенности

### Дизайн
- Минималистичный финтех-стиль
- Поддержка светлой/тёмной темы
- Адаптивный дизайн (desktop + mobile)
- Плавные анимации (Framer Motion)

### Функциональность
- Подключение банков (ABank, VBank, SBank)
- Агрегация счетов и транзакций
- Графики расходов и доходов
- Фильтрация транзакций
- Аналитика и прогнозы
- Управление подпиской (Free/Pro)

### Технические особенности
- Автоматический fallback на моки при ошибках API
- Skeleton loading states
- Error handling с понятными сообщениями
- PWA поддержка для установки на мобильные устройства
- TypeScript для типобезопасности

## 🔌 API интеграция

Приложение использует следующие endpoints:

- `POST /api/banks/{bank}/connect` - подключение банка
- `GET /api/accounts` - получение счетов
- `GET /api/transactions` - получение транзакций
- `GET /api/user/profile` - профиль пользователя
- `POST /api/user/subscribe` - оформление подписки

Все запросы включают заголовок `X-Request-Id` для трассировки.

## 📝 Примечание

Проект использует `pnpm` вместо `npm`. Убедитесь, что у вас установлен pnpm:

```bash
curl -fsSL https://get.pnpm.io/install.sh | sh -
source ~/.bashrc  # или перезапустите терминал
```

## 🎯 Навигация

- `/dashboard` - главный экран (по умолчанию)
- `/connect` - подключение банков
- `/transactions` - все транзакции
- `/analytics` - аналитика и прогнозы
- `/settings` - настройки
- `/subscription` - оформление подписки
- `/onboarding` - приветственный экран
