# Go: від початківця до практика

Структурована база знань по мові Go для повторення та закріплення.

## Розділи

### [Основи мови](./fundamentals/)

| Тема | Опис |
|------|------|
| [01 — Базові концепції](./fundamentals/01-basics/) | Змінні, типи, константи, zero values |
| [02 — Управління потоком](./fundamentals/02-control-flow/) | if, for, switch, defer |
| [03 — Функції](./fundamentals/03-functions/) | Функції, closures, variadic, defer |
| [04 — Типи та інтерфейси](./fundamentals/04-types/) | Structs, methods, interfaces, embedding |
| [05 — Конкурентність](./fundamentals/05-concurrency/) | Goroutines, channels, select, sync |
| [06 — Помилки](./fundamentals/06-errors/) | error, custom errors, panic/recover |
| [07 — Стандартна бібліотека](./fundamentals/07-stdlib/) | fmt, os, io, strings, time, net/http |

### [Патерни та техніки](./patterns/)

| Тема | Опис |
|------|------|
| [01 — Design Patterns](./patterns/01-design-patterns/) | Типові патерни в Go |
| [02 — Тестування](./patterns/02-testing/) | unit, table-driven tests, benchmarks |
| [03 — Продуктивність](./patterns/03-performance/) | profiling, оптимізації, pprof |

### [Веб-розробка](./web/)

| Фреймворк | Опис |
|-----------|------|
| [Echo](./web/echo/) | Швидкий мінімалістичний веб-фреймворк |
| [Chi](./web/chi/) | Легкий роутер, сумісний зі стандартним net/http |

### [Blockchain](./blockchain/)

| Тема | Опис |
|------|------|
| [Solana](./blockchain/solana/) | Solana програми та клієнти на Go |

---

## Як користуватись

- **Читати онлайн** — відкрий будь-яку директорію на GitHub, README.md відображається автоматично
- **Запустити локально** — `go run fundamentals/01-basics/basics.go`
- **Швидкий пошук** — `grep -r "goroutine" .`
