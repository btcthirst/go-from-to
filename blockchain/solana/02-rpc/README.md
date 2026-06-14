# 02 — RPC Client

## Підключення

```go
import "github.com/gagliardetto/solana-go/rpc"

client := rpc.New(rpc.DevNet_RPC)
// або з кастомним URL
client := rpc.New("https://my-rpc.example.com")
```

## Commitment рівні

```go
rpc.CommitmentProcessed  // найшвидший, може бути відкочений
rpc.CommitmentConfirmed  // підтверджено більшістю валідаторів
rpc.CommitmentFinalized  // фіналізовано, безпечно
```

## Основні методи

```go
// Мережа / статус
client.GetVersion(ctx)
client.GetSlot(ctx, rpc.CommitmentFinalized)
client.GetBlockHeight(ctx, rpc.CommitmentFinalized)
client.GetHealth(ctx)
client.GetMinimumBalanceForRentExemption(ctx, dataSize, commitment)

// Акаунти
client.GetBalance(ctx, pubkey, commitment)
client.GetAccountInfo(ctx, pubkey, &rpc.GetAccountInfoOpts{...})
client.GetMultipleAccounts(ctx, pubkeys...)
client.GetProgramAccounts(ctx, programID, &rpc.GetProgramAccountsOpts{...})

// Блоки / транзакції
client.GetRecentBlockhash(ctx, commitment)        // deprecated, але часто зустрічається
client.GetLatestBlockhash(ctx, commitment)        // актуальний
client.GetTransaction(ctx, sig, &opts)
client.GetSignaturesForAddress(ctx, pubkey, &opts)

// Airdrop (тільки devnet/testnet)
client.RequestAirdrop(ctx, pubkey, lamports, commitment)

// Токени
client.GetTokenAccountsByOwner(ctx, owner, &rpc.GetTokenAccountsConfig{...}, &opts)
client.GetTokenAccountBalance(ctx, tokenAccount, commitment)
```

## Помилки RPC

```go
// RPC повертає *jsonrpc.RPCError
var rpcErr *jsonrpc.RPCError
if errors.As(err, &rpcErr) {
    fmt.Println(rpcErr.Code, rpcErr.Message)
}
```

## Sysvar акаунти

Спеціальні акаунти з системною інформацією:

```
SysVarClockPubkey      — поточний slot, epoch, unix timestamp
SysVarRentPubkey       — параметри rent
SysVarEpochSchedulePubkey — розклад epochs
```
