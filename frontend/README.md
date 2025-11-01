# FinHelper Frontend

Frontend часть приложения для управления финансами, построенная на React + TypeScript с использованием Vite.

## Установка

```bash
pnpm install
```

## Настройка окружения

Создайте файл `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

Убедитесь, что `VITE_API_BASE_URL` указывает на ваш backend:

```
VITE_API_BASE_URL=http://localhost:8080
```

## Запуск разработки

```bash
pnpm run dev
```

Приложение будет доступно по адресу `http://localhost:5173`

## Сборка для продакшена

```bash
pnpm run build
```

## Предпросмотр продакшен сборки

```bash
pnpm run preview
```

## Линтинг

```bash
pnpm run lint
```

## Технологии

- **React 19** - UI библиотека
- **TypeScript** - типизация
- **Vite** - сборщик и dev-сервер
- **CSS** - стилизация
- **@tanstack/react-query** - управление серверным состоянием
- **react-router-dom** - роутинг
- **pnpm** - менеджер пакетов
- **ESLint** - линтер кода

## Структура проекта

```
frontend/
├── src/
│   ├── pages/          # Страницы приложения
│   │   ├── Connect.tsx       # Подключение банка
│   │   ├── Accounts.tsx      # Список счетов
│   │   └── Transactions.tsx # Операции
│   ├── hooks/          # React Query хуки
│   │   ├── useAccounts.ts
│   │   ├── useConnectBank.ts
│   │   └── useTransactions.ts
│   ├── lib/            # Утилиты
│   │   └── api.ts       # API клиент
│   ├── types/          # TypeScript типы
│   │   └── api.ts       # Типы API
│   ├── App.tsx         # Главный компонент с роутингом
│   ├── App.css         # Стили компонента
│   ├── main.tsx        # Точка входа
│   ├── index.css       # Глобальные стили
│   └── assets/         # Статические ресурсы
├── public/             # Публичные файлы
├── index.html          # HTML шаблон
├── package.json        # Зависимости и скрипты
├── .env.example        # Пример конфигурации окружения
├── tsconfig.json       # Настройки TypeScript
├── tsconfig.app.json   # TypeScript конфиг для приложения
├── tsconfig.node.json  # TypeScript конфиг для Node
└── vite.config.ts      # Конфигурация Vite
```

## API интеграция

Приложение ожидает следующие endpoints от backend:

- `POST /api/banks/{bank}/connect` - подключение банка (vbank, abank, sbank)
- `GET /api/accounts` - получение списка счетов
- `GET /api/transactions?from={date}&to={date}&bank={bank}` - получение операций

Все запросы включают заголовок `X-Request-Id` для трассировки.

## Примечание

Проект использует `pnpm` вместо `npm`. Убедитесь, что у вас установлен pnpm:

```bash
curl -fsSL https://get.pnpm.io/install.sh | sh -
source ~/.bashrc  # или перезапустите терминал
```
