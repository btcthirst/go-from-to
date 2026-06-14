# Solana на Go — gagliardetto/solana-go

## Встановлення

```bash
go get github.com/gagliardetto/solana-go
```

## Теми

| Тема | Запуск |
|------|--------|
| [01 — Keypair](./01-keypair/) | `go run ./01-keypair/` |
| [02 — RPC Client](./02-rpc/) | `go run ./02-rpc/` |
| [03 — Transactions](./03-transactions/) | `go run ./03-transactions/` |
| [04 — SPL Tokens](./04-spl-tokens/) | `go run ./04-spl-tokens/` |
| [05 — Programs & PDA](./05-programs/) | `go run ./05-programs/` |

## Кластери

```go
import "github.com/gagliardetto/solana-go/rpc"

rpc.DevNet_RPC     // https://api.devnet.solana.com
rpc.TestNet_RPC    // https://api.testnet.solana.com
rpc.MainNetBeta_RPC // https://api.mainnet-beta.solana.com
```

## Ключові типи

```go
solana.PublicKey      // [32]byte — адреса акаунту / програми
solana.PrivateKey     // []byte  — ed25519 private key
solana.Signature      // [64]byte — підпис транзакції
solana.Transaction    // транзакція з інструкціями
solana.Instruction    // одна інструкція (program + accounts + data)
```

## Отримати devnet SOL (airdrop)

```bash
solana airdrop 2 <PUBKEY> --url devnet
# або через RPC:
client.RequestAirdrop(ctx, pubkey, solana.LAMPORTS_PER_SOL, rpc.CommitmentFinalized)
```

## Одиниці

```
1 SOL = 1_000_000_000 lamports
solana.LAMPORTS_PER_SOL = 1_000_000_000
```
