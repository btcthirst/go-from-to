package main

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	ctx    = context.Background()
	client = rpc.New(rpc.DevNet_RPC)
)

func main() {
	findATA()
	parseMintAccount()
	parseTokenAccount()
	getTokenAccountsByOwner()
	buildCreateMintTx()
	buildMintToTx()
	buildTokenTransferTx()
}

// --- ATA: детермінована адреса токен-акаунту ---

func findATA() {
	fmt.Println("=== Associated Token Account (ATA) ===")

	owner := solana.MustPublicKeyFromBase58("7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs")
	// USDC mint на devnet (може відрізнятись)
	mint := solana.MustPublicKeyFromBase58("4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU")

	ata, _, err := solana.FindAssociatedTokenAddress(owner, mint)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}
	fmt.Printf("  Owner: %s\n", owner)
	fmt.Printf("  Mint:  %s\n", mint)
	fmt.Printf("  ATA:   %s\n", ata)

	// ATA — детермінована: той самий owner+mint завжди дає ту саму адресу
	ata2, _, _ := solana.FindAssociatedTokenAddress(owner, mint)
	fmt.Printf("  Deterministic: %t\n", ata.Equals(ata2))
}

// --- Парсинг Mint акаунту ---

func parseMintAccount() {
	fmt.Println("\n=== Parse Mint Account ===")

	// USDC devnet mint
	mintPubkey := solana.MustPublicKeyFromBase58("4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU")

	out, err := client.GetAccountInfo(ctx, mintPubkey)
	if err != nil || out.Value == nil {
		fmt.Println("  Не знайдено (mint може не існувати на devnet):", err)
		return
	}

	data := out.Value.Data.GetBinary()
	if len(data) < 44 {
		fmt.Println("  Недостатньо даних")
		return
	}

	// Mint layout: 4(opt discriminator)+32(mint_authority)+8(supply)+1(decimals)+1(is_init)+...
	// Спрощений парсинг для демонстрації
	supply := binary.LittleEndian.Uint64(data[36:44])
	decimals := data[44]
	isInit := data[45] != 0

	fmt.Printf("  Supply:   %d\n", supply)
	fmt.Printf("  Decimals: %d\n", decimals)
	fmt.Printf("  IsInit:   %t\n", isInit)
	fmt.Printf("  Owner:    %s\n", out.Value.Owner)
}

// --- Парсинг Token акаунту через SDK ---

func parseTokenAccount() {
	fmt.Println("\n=== Parse Token Account (SDK) ===")

	// Спробуємо знайти будь-який token account на devnet
	// Для демо — описуємо структуру
	fmt.Println("  Token Account layout (165 bytes):")
	fmt.Println("    [0:32]   mint       — який токен зберігається")
	fmt.Println("    [32:64]  owner      — власник акаунту")
	fmt.Println("    [64:72]  amount     — кількість токенів (u64, le)")
	fmt.Println("    [72:108] delegate   — Option<PublicKey>")
	fmt.Println("    [108]    state      — 0=uninitialized, 1=initialized, 2=frozen")
	fmt.Println("    [109:...]is_native, delegated_amount, close_authority")

	// Парсинг через SDK
	var acc token.Account
	fmt.Printf("\n  token.Account struct fields:\n")
	fmt.Printf("    Mint:    solana.PublicKey\n")
	fmt.Printf("    Owner:   solana.PublicKey\n")
	fmt.Printf("    Amount:  uint64\n")
	fmt.Printf("    State:   token.AccountState\n")
	_ = acc
}

// --- Токен-акаунти власника ---

func getTokenAccountsByOwner() {
	fmt.Println("\n=== GetTokenAccountsByOwner ===")

	owner := solana.MustPublicKeyFromBase58("7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs")

	out, err := client.GetTokenAccountsByOwner(
		ctx,
		owner,
		&rpc.GetTokenAccountsConfig{
			ProgramId: solana.TokenProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{
			Encoding: solana.EncodingBase64,
		},
	)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}

	fmt.Printf("  Token accounts: %d\n", len(out.Value))
	for i, ta := range out.Value {
		if i >= 3 {
			fmt.Println("  ...")
			break
		}
		fmt.Printf("  [%d] %s\n", i, ta.Pubkey)
	}
}

// --- Побудова транзакцій (без відправки) ---

func buildCreateMintTx() {
	fmt.Println("\n=== Build: Create Mint + ATA ===")

	payer := solana.NewWallet()
	mintKP := solana.NewWallet() // окремий keypair для mint акаунту
	owner := payer.PublicKey()

	// Rent для mint акаунту (82 bytes)
	mintRent, err := client.GetMinimumBalanceForRentExemption(ctx, 82, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  GetMinimumBalanceForRentExemption error:", err)
		return
	}

	// Rent для token акаунту (165 bytes)
	tokenRent, err := client.GetMinimumBalanceForRentExemption(ctx, 165, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}

	// ATA адреса
	ata, _, _ := solana.FindAssociatedTokenAddress(owner, mintKP.PublicKey())

	// Інструкції
	instructions := []solana.Instruction{
		// 1. Створити системний акаунт для mint
		solana.NewInstruction(
			solana.SystemProgramID,
			solana.AccountMetaSlice{
				{PublicKey: payer.PublicKey(), IsSigner: true, IsWritable: true},
				{PublicKey: mintKP.PublicKey(), IsSigner: true, IsWritable: true},
			},
			// createAccount instruction data: discriminator(4) + lamports(8) + space(8) + programId(32)
			func() []byte {
				data := make([]byte, 52)
				binary.LittleEndian.PutUint32(data[0:], 0)              // createAccount
				binary.LittleEndian.PutUint64(data[4:], mintRent)       // lamports
				binary.LittleEndian.PutUint64(data[12:], 82)            // space
				copy(data[20:], solana.TokenProgramID[:])                // owner
				return data
			}(),
		),
		// 2. Ініціалізувати mint (6 decimals)
		// freezeAuthority = ZeroKey означає відключену заморозку
		token.NewInitializeMint2Instruction(
			6,                      // decimals
			owner,                  // mintAuthority
			solana.PublicKey{},     // freezeAuthority (zero = відключена)
			mintKP.PublicKey(),
		).Build(),
		// 3. Створити ATA для payer
		associatedtokenaccount.NewCreateInstruction(
			payer.PublicKey(), // payer
			owner,             // wallet
			mintKP.PublicKey(), // mint
		).Build(),
	}

	bhOut, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  blockhash error:", err)
		return
	}

	tx, err := solana.NewTransaction(instructions, bhOut.Value.Blockhash,
		solana.TransactionPayer(payer.PublicKey()),
	)
	if err != nil {
		fmt.Println("  build error:", err)
		return
	}

	fmt.Printf("  Mint:         %s\n", mintKP.PublicKey())
	fmt.Printf("  ATA:          %s\n", ata)
	fmt.Printf("  MintRent:     %d lamports\n", mintRent)
	fmt.Printf("  TokenRent:    %d lamports\n", tokenRent)
	fmt.Printf("  Instructions: %d\n", len(tx.Message.Instructions))
	fmt.Printf("  Signers needed: payer + mint keypair\n")
}

func buildMintToTx() {
	fmt.Println("\n=== Build: MintTo ===")

	payer := solana.NewWallet()
	mint := solana.NewWallet().PublicKey()
	destination, _, _ := solana.FindAssociatedTokenAddress(payer.PublicKey(), mint)

	decimals := uint8(6)
	amount := uint64(1000) * uint64(1e6) // 1000 токенів (6 decimals)

	instr := token.NewMintToInstruction(
		amount,
		mint,
		destination,
		payer.PublicKey(), // mint authority
		[]solana.PublicKey{},
	).Build()

	fmt.Printf("  Mint:        %s\n", mint)
	fmt.Printf("  Destination: %s\n", destination)
	fmt.Printf("  Amount:      %d (raw) = %d tokens (%d decimals)\n", amount, amount/uint64(1e6), decimals)
	fmt.Printf("  Instruction: %v\n", instr.ProgramID())
}

func buildTokenTransferTx() {
	fmt.Println("\n=== Build: Token Transfer ===")

	sender := solana.NewWallet()
	receiver := solana.NewWallet()
	mint := solana.NewWallet().PublicKey()

	senderATA, _, _ := solana.FindAssociatedTokenAddress(sender.PublicKey(), mint)
	receiverATA, _, _ := solana.FindAssociatedTokenAddress(receiver.PublicKey(), mint)

	amount := uint64(100) * uint64(1e6) // 100 токенів

	instructions := []solana.Instruction{
		// Створити ATA для receiver якщо не існує
		associatedtokenaccount.NewCreateIdempotentInstruction(
			sender.PublicKey(), // payer
			receiver.PublicKey(),
			mint,
		).Build(),
		// Transfer
		token.NewTransferInstruction(
			amount,
			senderATA,
			receiverATA,
			sender.PublicKey(), // owner
			[]solana.PublicKey{},
		).Build(),
	}

	fmt.Printf("  Sender ATA:   %s\n", senderATA)
	fmt.Printf("  Receiver ATA: %s\n", receiverATA)
	fmt.Printf("  Amount:       %d tokens\n", amount/uint64(1e6))
	fmt.Printf("  Instructions: %d\n", len(instructions))
	fmt.Println("  (CreateIdempotent — безпечно: не падає якщо ATA вже існує)")
}
