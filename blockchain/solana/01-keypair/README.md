# 01 — Keypair

Keypair у Solana — це пара ed25519 ключів: приватний (64 байти) та публічний (32 байти).

## Типи

```go
solana.PrivateKey  // []byte, 64 bytes (seed + pubkey)
solana.PublicKey   // [32]byte
```

## Генерація

```go
// Новий випадковий keypair
wallet := solana.NewWallet()
wallet.PrivateKey   // solana.PrivateKey
wallet.PublicKey()  // solana.PublicKey
```

## Імпорт

```go
// З base58 (як у Phantom export)
pk := solana.MustPrivateKeyFromBase58("5HueCG...")

// З JSON масиву байтів (як у Solana CLI keypair.json)
pk, err := solana.PrivateKeyFromSolanaKeygenFile("~/.config/solana/id.json")

// З seed phrase (BIP39)
seed := bip39.NewSeed(mnemonic, "")
pk := solana.PrivateKey(ed25519.NewKeyFromSeed(seed[:32]))
```

## Підпис повідомлень

```go
msg := []byte("hello solana")
sig, err := pk.Sign(msg)

ok, err := pk.PublicKey().Verify(msg, sig)
```

## Публічний ключ

```go
pubkey := pk.PublicKey()
pubkey.String()       // base58 рядок
pubkey.ToPointer()    // *solana.PublicKey
pubkey.IsZero()       // чи нульовий
pubkey.Equals(other)  // порівняння
```
