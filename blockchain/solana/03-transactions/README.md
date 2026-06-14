# 03 — Transactions

## Анатомія транзакції

```
Transaction
├── Signatures        []Signature          — підписи (по одному на кожен signer)
└── Message
    ├── Header        — кількість signers/readonly акаунтів
    ├── AccountKeys   []PublicKey          — всі акаунти транзакції
    ├── RecentBlockhash                    — TTL транзакції (~150 слотів)
    └── Instructions  []CompiledInstruction
            ├── ProgramIDIndex  — індекс у AccountKeys
            ├── Accounts        — індекси акаунтів у AccountKeys
            └── Data            — бінарні дані для програми
```

## Lifecycle

```
1. Отримати свіжий blockhash
2. Побудувати інструкції
3. Зібрати транзакцію
4. Підписати
5. Симулювати (опційно)
6. Відправити
7. Підтвердити
```

## Побудова транзакції

```go
// Через TransactionBuilder (зручно)
tx, err := solana.NewTransaction(
    []solana.Instruction{instr1, instr2},
    recentBlockhash,
    solana.TransactionPayer(payer.PublicKey()),
)

// Підписати
_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
    if key.Equals(payer.PublicKey()) {
        return &payer.PrivateKey
    }
    return nil
})
```

## Відправка та підтвердження

```go
sig, err := client.SendTransactionWithOpts(ctx, tx,
    rpc.TransactionOpts{
        SkipPreflight:       false,    // виконати preflight simulation
        PreflightCommitment: rpc.CommitmentProcessed,
    },
)

// Підтвердити (polling)
_, err = confirm.WaitForConfirmation(ctx, client, sig, rpc.CommitmentConfirmed)
```

## Preflight simulation

```go
out, err := client.SimulateTransaction(ctx, tx)
if out.Value.Err != nil {
    fmt.Println("simulation failed:", out.Value.Err)
    fmt.Println("logs:", out.Value.Logs)
}
```

## Fees

```go
// Вартість транзакції = 5000 lamports * кількість підписів
// + compute units (якщо використовуються)
out, err := client.GetFeeForMessage(ctx, tx.Message, rpc.CommitmentFinalized)
```

## Версіонні транзакції (v0)

```go
// v0 підтримує Address Lookup Tables (зменшує розмір tx)
// Стандартний SDK поки працює з legacy транзакціями
```
