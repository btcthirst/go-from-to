# Тестування проекту

## ✅ Статус тестування

```
✅ internal/app                        - OK (41.0% покриття)
✅ internal/config                     - OK (84.4% покриття)
✅ internal/excel/reader               - OK (6.5% покриття*)
```

*Покриття низька для reader тому що більшість кода - це читання файлів, які складно тестувати в одиничних тестах

## Швидкий старт

### Запуск всіх тестів
```bash
make test
```

### Запуск тестів з виводом
```bash
make test-verbose
```

### Запуск швидких тестів
```bash
make test-short
```

### Генерування звіту покриття
```bash
make test-coverage
```

## Статистика тестів

| Пакет | Тести | Бенчмарки | Статус |
|-------|-------|-----------|--------|
| app | 8 | 1 | ✅ PASS |
| config | 6 | 2 | ✅ PASS |
| reader | 2 | 2 | ✅ PASS |
| **ВСЬОГО** | **16** | **5** | **✅ PASS** |

## Результати бенчмарків

### BenchmarkFindAccountInCounterparty (10,000 рахунків)
```
7245 iterations  162559 ns/op  0 B/op (0 алокацій)
~162 мікросекунди на операцію
```

### BenchmarkIsValidName
```
Дуже швидко (~100-200 ns/op)
```

### BenchmarkProcessPayments (1,000 платежів, 500 рахунків)
```
Обробляється за мілісекунди
```

## Тестові кейси

### app_test.go (8 тестів)
- ✅ Basic payment processing
- ✅ Empty payments handling
- ✅ Empty account map
- ✅ Consistent data preservation
- ✅ Account finding
- ✅ Integration: Full workflow
- ✅ Integration: Large payment batch (10,000)
- ✅ Integration: Deduction classification

### config_test.go (6 тестів + 7 кейсів)
- ✅ Load configuration from YAML
- ✅ Load non-existent file error
- ✅ Get document config
- ✅ Get non-existent config error
- ✅ Get config without loading error
- ✅ Get column index
- ✅ Validate row

### reader_test.go (2 функції + 10 кейсів)
- ✅ Find account in counterparty (6 кейсів)
- ✅ Valid name validation (9 кейсів)

## Покриття коду

```
internal/app          41.0%  (можна підвищити тестами UI)
internal/config       84.4%  (дуже гарне)
internal/excel/reader 6.5%   (тяжко тестувати файли)
```

## CI/CD Інтеграція

Для CI/CD pipeline рекомендується:

```bash
# Швидкі тести (без інтеграційних)
go test -short ./...

# Всі тести
go test ./...

# Спеціальні перевірки
go test -race ./...          # Перевірка race conditions
go test -coverprofile=c.out ./...
```
