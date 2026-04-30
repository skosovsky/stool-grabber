## stool-grabber

CLI-утилита для скрапинга обсуждений Telegram-канала (комментарии), агрегации по пользователям и опционального LLM-анализа с генерацией Markdown-отчёта.

### Запуск

#### 1) Подготовьте переменные окружения

Переменные можно задать в окружении процесса, либо создать файл `.env` (можно скопировать из `.env.example`) — CLI подхватит его автоматически при запуске.

- **Telegram MTProto**:
  - `TG_APP_ID` — integer (из `https://my.telegram.org`)
  - `TG_APP_HASH` — app hash
  - `TG_SESSION_PATH` — путь к файлу сессии (опционально, по умолчанию `session.json`; относительные пути будут нормализованы до абсолютных)
- **OpenRouter**:
  - `OPENROUTER_API_KEY` — ключ API
  - `OPENROUTER_BASE_URL` — опционально, по умолчанию `https://openrouter.ai/api/v1`

#### 2) Подготовьте `job.yaml`

Пример минимального конфига:

```yaml
version: "1"
target:
  channel_username: "@your_channel"
  parse_depth: 20
  delay_ms: 200
filter:
  min_comments_to_analyze: 2
  max_users_to_analyze: 20
  exclude_admins: true
llm:
  provider: "openrouter"
  model: "openai/gpt-4o-mini"
  temperature: 0.2
  prompt_template: |
    Кратко опиши самых активных участников и горячие темы.
output:
  format: "markdown"
  filepath: "./out/report.md"
```

#### 3) Запуск

```bash
go run ./cmd/stool-grabber run -c job.yaml
```

### Manual smoke-check (Task 10)

1. **Первый запуск**:
   - CLI должен попросить интерактивную авторизацию (телефон/код/2FA при необходимости).
   - Должен появиться файл сессии по `TG_SESSION_PATH` (или `session.json`, если переменная не задана).
2. **Второй запуск**:
   - Авторизация не должна запрашиваться повторно (сессия переиспользуется).
3. **Short-circuit**:
   - Если после фильтров `min_comments_to_analyze`/`exclude_admins` и top-N список пользователей пуст, **OpenRouter не должен вызываться**.
4. **Отчёт**:
   - Markdown печатается в stdout.
   - Если `output.filepath` задан — файл создаётся/перезаписывается.

### Проверки

```bash
go test ./...
```

