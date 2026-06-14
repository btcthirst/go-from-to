# 04 — SPL Tokens

## Концепції

```
Mint Account       — описує токен: supply, decimals, mint authority
Token Account      — зберігає баланс конкретного токену для конкретного власника
Associated Token   — стандартна адреса: ATA(owner, mint) → детермінована
Account (ATA)
```

## Адреса ATA

```go
import "github.com/gagliardetto/solana-go/programs/associated-token-account"

ata, _, err := solana.FindAssociatedTokenAddress(
    owner,       // solana.PublicKey
    mint,        // solana.PublicKey
)
```

## Структура Mint акаунту

```
MintAccount (82 bytes):
  mint_authority     Option<PublicKey>  — хто може емітувати нові токени
  supply             u64               — загальна кількість токенів
  decimals           u8                — кількість знаків після коми
  is_initialized     bool
  freeze_authority   Option<PublicKey>  — хто може заморозити token accounts
```

## Основні інструкції Token Program

```go
// Створити mint акаунт (2 кроки: SystemProgram + InitializeMint)
token.NewInitializeMintInstruction(decimals, mintAuthority, freezeAuthority, mint, sysvar.Rent)

// Створити token account
token.NewInitializeAccountInstruction(account, mint, owner, sysvar.Rent)

// Або через ATA (рекомендовано)
associatedtokenaccount.NewCreateInstruction(payer, owner, mint)

// Mint (емісія)
token.NewMintToInstruction(amount, mint, destination, mintAuthority, []solana.PublicKey{})

// Transfer
token.NewTransferInstruction(amount, source, destination, owner, []solana.PublicKey{})

// Burn
token.NewBurnInstruction(amount, account, mint, owner, []solana.PublicKey{})

// Close (повернути rent)
token.NewCloseAccountInstruction(account, destination, owner, []solana.PublicKey{})
```

## Rent exemption

Token account потребує мінімум lamports щоб існувати:

```go
// Mint account: 82 bytes → ~1.46 mSOL
// Token account: 165 bytes → ~2.04 mSOL
rent, _ := client.GetMinimumBalanceForRentExemption(ctx, 165, rpc.CommitmentFinalized)
```

## Decimals та amounts

```
decimals = 6  → 1 токен = 1_000_000 (як USDC)
decimals = 9  → 1 токен = 1_000_000_000 (як SOL)
decimals = 0  → NFT-style (тільки цілі числа)
```
