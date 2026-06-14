package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gagliardetto/solana-go"
	"github.com/tyler-smith/go-bip39"
)

func main() {
	generateKeypair()
	importFromBase58()
	importFromFile()
	fromMnemonic()
	signAndVerify()
	publicKeyOps()
}

// --- Генерація нового keypair ---

func generateKeypair() {
	fmt.Println("=== Генерація keypair ===")

	wallet := solana.NewWallet()
	fmt.Println("  PublicKey: ", wallet.PublicKey())
	fmt.Println("  PrivateKey:", wallet.PrivateKey.String())

	// Зберегти у файл (формат Solana CLI: JSON масив байтів)
	path := "/tmp/demo-keypair.json"
	if err := saveKeypair(wallet.PrivateKey, path); err != nil {
		fmt.Println("  saveKeypair error:", err)
		return
	}
	fmt.Println("  Збережено:", path)
}

func saveKeypair(pk solana.PrivateKey, path string) error {
	bytes := []byte(pk)
	data, err := json.Marshal(bytes)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// --- Імпорт з base58 ---

func importFromBase58() {
	fmt.Println("\n=== Імпорт з base58 ===")

	// Генеруємо новий і одразу імпортуємо назад (демо)
	original := solana.NewWallet()
	b58 := original.PrivateKey.String()
	fmt.Println("  base58:", b58[:20]+"...")

	pk := solana.MustPrivateKeyFromBase58(b58)
	fmt.Println("  PublicKey:", pk.PublicKey())
	fmt.Println("  Match:", pk.PublicKey().Equals(original.PublicKey()))
}

// --- Імпорт з JSON файлу ---

func importFromFile() {
	fmt.Println("\n=== Імпорт з JSON файлу ===")

	path := "/tmp/demo-keypair.json"
	pk, err := loadKeypair(path)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}
	fmt.Println("  PublicKey:", pk.PublicKey())
}

func loadKeypair(path string) (solana.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bytes []byte
	if err := json.Unmarshal(data, &bytes); err != nil {
		return nil, err
	}
	return solana.PrivateKey(bytes), nil
}

// --- З мнемонічної фрази (BIP39) ---

func fromMnemonic() {
	fmt.Println("\n=== Keypair з мнемоніки (BIP39) ===")

	// Генеруємо нову мнемоніку
	entropy, _ := bip39.NewEntropy(128) // 12 слів
	mnemonic, _ := bip39.NewMnemonic(entropy)
	fmt.Println("  Mnemonic:", mnemonic)

	// Seed → ed25519 → Solana keypair
	seed := bip39.NewSeed(mnemonic, "") // passphrase = ""
	edKey := ed25519.NewKeyFromSeed(seed[:32])
	pk := solana.PrivateKey(edKey)

	fmt.Println("  PublicKey:", pk.PublicKey())

	// Детермінований: та сама мнемоніка → той самий keypair
	seed2 := bip39.NewSeed(mnemonic, "")
	pk2 := solana.PrivateKey(ed25519.NewKeyFromSeed(seed2[:32]))
	fmt.Println("  Deterministic:", pk.PublicKey().Equals(pk2.PublicKey()))
}

// --- Підпис і верифікація ---

func signAndVerify() {
	fmt.Println("\n=== Підпис повідомлень ===")

	wallet := solana.NewWallet()
	msg := []byte("Hello, Solana!")

	sig, err := wallet.PrivateKey.Sign(msg)
	if err != nil {
		fmt.Println("  sign error:", err)
		return
	}
	fmt.Println("  Signature:", sig)

	ok := wallet.PublicKey().Verify(msg, sig)
	fmt.Printf("  Verify (valid msg): %t\n", ok)

	// Неправильне повідомлення
	ok = wallet.PublicKey().Verify([]byte("tampered!"), sig)
	fmt.Printf("  Verify (tampered):  %t\n", ok)

	// Чужий pubkey
	other := solana.NewWallet()
	ok = other.PublicKey().Verify(msg, sig)
	fmt.Printf("  Verify (wrong key): %t\n", ok)
}

// --- Операції з PublicKey ---

func publicKeyOps() {
	fmt.Println("\n=== PublicKey операції ===")

	w := solana.NewWallet()
	pk := w.PublicKey()

	fmt.Println("  String (base58):", pk.String())
	fmt.Println("  Bytes:", pk[:8], "...")
	fmt.Println("  IsZero:", pk.IsZero())

	zero := solana.PublicKey{}
	fmt.Println("  Zero key IsZero:", zero.IsZero())

	// Парсинг з рядка
	parsed, err := solana.PublicKeyFromBase58(pk.String())
	if err != nil {
		fmt.Println("  parse error:", err)
		return
	}
	fmt.Println("  Parsed == Original:", parsed.Equals(pk))

	// Відомі програми
	fmt.Println("\n  Системні програми:")
	fmt.Println("  SystemProgram:     ", solana.SystemProgramID)
	fmt.Println("  TokenProgram:      ", solana.TokenProgramID)
	fmt.Println("  Token2022Program:  ", solana.Token2022ProgramID)
	fmt.Println("  AssociatedToken:   ", solana.SPLAssociatedTokenAccountProgramID)
	fmt.Println("  SysvarRent:        ", solana.SysVarRentPubkey)
	fmt.Println("  SysvarClock:       ", solana.SysVarClockPubkey)
}
