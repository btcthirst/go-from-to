package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	ctx    = context.Background()
	client = rpc.New(rpc.DevNet_RPC)
)

func main() {
	pdaDerivation()
	ataPDA()
	metaplexPDA()
	anchorDiscriminator()
	buildCustomInstruction()
	inspectProgram()
}

// --- PDA деривація ---

func pdaDerivation() {
	fmt.Println("=== PDA Derivation ===")

	programID := solana.MustPublicKeyFromBase58("11111111111111111111111111111112") // SystemProgram — для демо
	owner := solana.NewWallet().PublicKey()

	// Одне seed
	pda1, bump1, err := solana.FindProgramAddress(
		[][]byte{[]byte("vault")},
		programID,
	)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}
	fmt.Printf("  seed=[\"vault\"]:         %s (bump=%d)\n", pda1, bump1)

	// Кілька seeds
	pda2, bump2, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte("user"),
			owner.Bytes(),
		},
		programID,
	)
	fmt.Printf("  seed=[\"user\", owner]:   %s (bump=%d)\n", pda2, bump2)

	// З числовим seed
	idBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBytes, 42)
	pda3, bump3, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte("order"),
			idBytes,
		},
		programID,
	)
	fmt.Printf("  seed=[\"order\", 42]:    %s (bump=%d)\n", pda3, bump3)

	// Детермінований: той самий вхід → той самий PDA
	pda1b, bump1b, _ := solana.FindProgramAddress([][]byte{[]byte("vault")}, programID)
	fmt.Printf("  Deterministic:         %t (bump=%d)\n", pda1.Equals(pda1b), bump1b)
}

// --- ATA PDA (як це працює всередині) ---

func ataPDA() {
	fmt.Println("\n=== ATA PDA (внутрішня механіка) ===")

	owner := solana.NewWallet().PublicKey()
	mint := solana.NewWallet().PublicKey()

	// ATA derivation: seeds = [owner, token_program, mint]
	manual, _, _ := solana.FindProgramAddress(
		[][]byte{
			owner.Bytes(),
			solana.TokenProgramID.Bytes(),
			mint.Bytes(),
		},
		solana.SPLAssociatedTokenAccountProgramID,
	)

	// Офіційна функція
	official, _, _ := solana.FindAssociatedTokenAddress(owner, mint)

	fmt.Printf("  Owner:    %s\n", owner)
	fmt.Printf("  Mint:     %s\n", mint)
	fmt.Printf("  Manual:   %s\n", manual)
	fmt.Printf("  Official: %s\n", official)
	fmt.Printf("  Match:    %t\n", manual.Equals(official))
}

// --- Metaplex Metadata PDA ---

func metaplexPDA() {
	fmt.Println("\n=== Metaplex PDA ===")

	metaplexID := solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
	mint := solana.NewWallet().PublicKey()

	// Metadata PDA
	metadataPDA, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte("metadata"),
			metaplexID.Bytes(),
			mint.Bytes(),
		},
		metaplexID,
	)

	// Edition PDA
	editionPDA, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte("metadata"),
			metaplexID.Bytes(),
			mint.Bytes(),
			[]byte("edition"),
		},
		metaplexID,
	)

	fmt.Printf("  Mint:         %s\n", mint)
	fmt.Printf("  Metadata PDA: %s\n", metadataPDA)
	fmt.Printf("  Edition PDA:  %s\n", editionPDA)
}

// --- Anchor discriminator ---

func anchorDiscriminator() {
	fmt.Println("\n=== Anchor Discriminator ===")

	// Anchor: discriminator = sha256("global:<instruction_name>")[:8]
	instructions := []string{"initialize", "deposit", "withdraw", "transfer"}

	for _, name := range instructions {
		h := sha256.Sum256([]byte("global:" + name))
		disc := h[:8]
		fmt.Printf("  %-12s → [%d %d %d %d %d %d %d %d]\n",
			name, disc[0], disc[1], disc[2], disc[3], disc[4], disc[5], disc[6], disc[7])
	}

	// Account discriminator = sha256("account:<AccountName>")[:8]
	accounts := []string{"UserState", "VaultAccount", "PoolConfig"}
	fmt.Println()
	for _, name := range accounts {
		h := sha256.Sum256([]byte("account:" + name))
		disc := h[:8]
		fmt.Printf("  %-14s → [%d %d %d %d %d %d %d %d]\n",
			name, disc[0], disc[1], disc[2], disc[3], disc[4], disc[5], disc[6], disc[7])
	}
}

// --- Побудова кастомної інструкції ---

func buildCustomInstruction() {
	fmt.Println("\n=== Custom Instruction ===")

	// Уявна програма
	programID := solana.NewWallet().PublicKey()
	payer := solana.NewWallet()
	pdaSeeds := [][]byte{[]byte("config"), payer.PublicKey().Bytes()}
	configPDA, bump, _ := solana.FindProgramAddress(pdaSeeds, programID)

	// Instruction data: discriminator(8) + amount(8) + bump(1)
	disc := sha256.Sum256([]byte("global:initialize"))
	data := make([]byte, 17)
	copy(data[0:8], disc[:8])
	binary.LittleEndian.PutUint64(data[8:16], 1_000_000)
	data[16] = bump

	instr := solana.NewInstruction(
		programID,
		solana.AccountMetaSlice{
			solana.Meta(payer.PublicKey()).SIGNER().WRITE(), // payer
			solana.Meta(configPDA).WRITE(),                 // PDA account (не підписант)
			solana.Meta(solana.SystemProgramID),             // system program для створення
		},
		data,
	)

	fmt.Printf("  Program:      %s\n", programID)
	fmt.Printf("  PDA:          %s (bump=%d)\n", configPDA, bump)
	instrData, _ := instr.Data()
	fmt.Printf("  Data length:  %d bytes\n", len(instrData))
	fmt.Printf("  Accounts:     %d\n", len(instr.Accounts()))
	for i, acc := range instr.Accounts() {
		fmt.Printf("    [%d] %s signer=%t writable=%t\n",
			i, acc.PublicKey, acc.IsSigner, acc.IsWritable)
	}
}

// --- Інспекція програми на devnet ---

func inspectProgram() {
	fmt.Println("\n=== Inspect Program ===")

	// Token Program
	programs := map[string]solana.PublicKey{
		"SystemProgram":   solana.SystemProgramID,
		"TokenProgram":    solana.TokenProgramID,
		"Token2022":       solana.Token2022ProgramID,
		"AssociatedToken": solana.SPLAssociatedTokenAccountProgramID,
	}

	for name, pk := range programs {
		out, err := client.GetAccountInfo(ctx, pk)
		if err != nil || out.Value == nil {
			fmt.Printf("  %-20s — error: %v\n", name, err)
			continue
		}
		acc := out.Value
		fmt.Printf("  %-20s — executable=%t, owner=%s, data=%d bytes\n",
			name, acc.Executable, acc.Owner, len(acc.Data.GetBinary()))
	}
}
