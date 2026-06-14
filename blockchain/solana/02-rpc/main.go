package main

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	ctx    = context.Background()
	client = rpc.New(rpc.DevNet_RPC)
)

func main() {
	networkInfo()
	getBalance()
	getAccountInfo()
	getLatestBlockhash()
	getTransactionHistory()
	getSysvar()
}

// --- Інформація про мережу ---

func networkInfo() {
	fmt.Println("=== Мережа ===")

	version, err := client.GetVersion(ctx)
	if err != nil {
		fmt.Println("  GetVersion error:", err)
	} else {
		fmt.Println("  Version:", version.SolanaCore)
	}

	slot, err := client.GetSlot(ctx, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  GetSlot error:", err)
	} else {
		fmt.Println("  Slot:", slot)
	}

	height, err := client.GetBlockHeight(ctx, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  GetBlockHeight error:", err)
	} else {
		fmt.Println("  BlockHeight:", height)
	}
}

// --- Баланс акаунту ---

func getBalance() {
	fmt.Println("\n=== GetBalance ===")

	// Використаємо відомий публічний акаунт (Solana Foundation)
	pubkey := solana.MustPublicKeyFromBase58("7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs")

	out, err := client.GetBalance(ctx, pubkey, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}

	lamports := out.Value
	sol := float64(lamports) / float64(solana.LAMPORTS_PER_SOL)
	fmt.Printf("  %s\n", pubkey)
	fmt.Printf("  Balance: %d lamports (%.9f SOL)\n", lamports, sol)
	fmt.Printf("  Context slot: %d\n", out.Context.Slot)
}

// --- Інформація про акаунт ---

func getAccountInfo() {
	fmt.Println("\n=== GetAccountInfo ===")

	// SystemProgram — завжди існує
	pubkey := solana.SystemProgramID

	out, err := client.GetAccountInfo(ctx, pubkey)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}

	if out.Value == nil {
		fmt.Println("  account not found")
		return
	}

	acc := out.Value
	fmt.Printf("  Owner:      %s\n", acc.Owner)
	fmt.Printf("  Lamports:   %d\n", acc.Lamports)
	fmt.Printf("  Executable: %t\n", acc.Executable)
	fmt.Printf("  Data len:   %d bytes\n", len(acc.Data.GetBinary()))
	fmt.Printf("  RentEpoch:  %d\n", acc.RentEpoch)
}

// --- Latest blockhash ---

func getLatestBlockhash() {
	fmt.Println("\n=== GetLatestBlockhash ===")

	out, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}

	fmt.Printf("  Blockhash:            %s\n", out.Value.Blockhash)
	fmt.Printf("  LastValidBlockHeight: %d\n", out.Value.LastValidBlockHeight)
	fmt.Printf("  Context slot:         %d\n", out.Context.Slot)
}

// --- Історія транзакцій ---

func getTransactionHistory() {
	fmt.Println("\n=== GetSignaturesForAddress ===")

	// Популярний devnet акаунт
	pubkey := solana.MustPublicKeyFromBase58("7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs")

	limit := 3
	out, err := client.GetSignaturesForAddressWithOpts(
		ctx,
		pubkey,
		&rpc.GetSignaturesForAddressOpts{
			Limit:      &limit,
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}

	fmt.Printf("  Останні %d транзакції:\n", len(out))
	for _, s := range out {
		status := "ok"
		if s.Err != nil {
			status = fmt.Sprintf("err: %v", s.Err)
		}
		fmt.Printf("  [slot %d] %s — %s\n", s.Slot, s.Signature, status)
	}
}

// --- Sysvar акаунти ---

func getSysvar() {
	fmt.Println("\n=== Sysvar: Clock ===")

	out, err := client.GetAccountInfo(ctx, solana.SysVarClockPubkey)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}
	if out.Value == nil {
		fmt.Println("  not found")
		return
	}

	// Clock sysvar layout: slot(8) + epoch_start_timestamp(8) + epoch(8) + leader_schedule_epoch(8) + unix_timestamp(8)
	data := out.Value.Data.GetBinary()
	if len(data) < 40 {
		fmt.Println("  unexpected data length")
		return
	}

	readU64 := func(b []byte, offset int) uint64 {
		var v uint64
		for i := range 8 {
			v |= uint64(b[offset+i]) << (8 * i)
		}
		return v
	}

	slot := readU64(data, 0)
	epoch := readU64(data, 16)
	unixTs := int64(readU64(data, 32))

	fmt.Printf("  Slot:           %d\n", slot)
	fmt.Printf("  Epoch:          %d\n", epoch)
	fmt.Printf("  Unix timestamp: %d\n", unixTs)
}
