package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	ctx    = context.Background()
	client = rpc.New(rpc.DevNet_RPC)
)

func main() {
	buildAndInspect()
	simulateTransfer()
	// sendTransfer()  // розкоментуй коли є funded акаунт
}

// --- Побудова та інспекція транзакції (без відправки) ---

func buildAndInspect() {
	fmt.Println("=== Побудова транзакції ===")

	sender := solana.NewWallet()
	receiver := solana.NewWallet()

	// Заглушка blockhash для демонстрації структури
	blockhash := solana.MustHashFromBase58("4vJ9JU1bJJE96FWSJKvHsmmFADCg4gpZQff4P3bkLKi")

	// Побудова SOL transfer інструкції
	amount := uint64(0.01 * float64(solana.LAMPORTS_PER_SOL)) // 0.01 SOL
	instruction := system.NewTransferInstruction(
		amount,
		sender.PublicKey(),
		receiver.PublicKey(),
	).Build()

	// Збираємо транзакцію
	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		blockhash,
		solana.TransactionPayer(sender.PublicKey()),
	)
	if err != nil {
		fmt.Println("  NewTransaction error:", err)
		return
	}

	// Підписуємо
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(sender.PublicKey()) {
			return &sender.PrivateKey
		}
		return nil
	})
	if err != nil {
		fmt.Println("  Sign error:", err)
		return
	}

	// Інспекція
	fmt.Printf("  Sender:       %s\n", sender.PublicKey())
	fmt.Printf("  Receiver:     %s\n", receiver.PublicKey())
	fmt.Printf("  Amount:       %d lamports (%.4f SOL)\n", amount, float64(amount)/float64(solana.LAMPORTS_PER_SOL))
	fmt.Printf("  Signatures:   %d\n", len(tx.Signatures))
	fmt.Printf("  Signature:    %s\n", tx.Signatures[0])
	fmt.Printf("  Instructions: %d\n", len(tx.Message.Instructions))
	fmt.Printf("  AccountKeys:  %d\n", len(tx.Message.AccountKeys))
	for i, k := range tx.Message.AccountKeys {
		fmt.Printf("    [%d] %s\n", i, k)
	}

	// Серіалізація
	encoded, err := tx.MarshalBinary()
	if err != nil {
		fmt.Println("  MarshalBinary error:", err)
		return
	}
	fmt.Printf("  Serialized:   %d bytes\n", len(encoded))
}

// --- Simulation (без реальних грошей) ---

func simulateTransfer() {
	fmt.Println("\n=== Simulation ===")

	sender := solana.NewWallet()
	receiver := solana.NewWallet()

	// Отримуємо реальний blockhash з devnet
	bhOut, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  GetLatestBlockhash error:", err)
		return
	}
	blockhash := bhOut.Value.Blockhash

	instruction := system.NewTransferInstruction(
		solana.LAMPORTS_PER_SOL/100, // 0.01 SOL
		sender.PublicKey(),
		receiver.PublicKey(),
	).Build()

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		blockhash,
		solana.TransactionPayer(sender.PublicKey()),
	)
	if err != nil {
		fmt.Println("  NewTransaction error:", err)
		return
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(sender.PublicKey()) {
			return &sender.PrivateKey
		}
		return nil
	})
	if err != nil {
		fmt.Println("  Sign error:", err)
		return
	}

	// Симуляція — транзакція перевіряється без виконання
	simResult, err := client.SimulateTransaction(ctx, tx)
	if err != nil {
		fmt.Println("  SimulateTransaction error:", err)
		return
	}

	if simResult.Value.Err != nil {
		// Очікувано: акаунт не фінансований
		fmt.Printf("  Simulation err (очікувано): %v\n", simResult.Value.Err)
	} else {
		fmt.Println("  Simulation: success")
	}

	fmt.Println("  Logs:")
	for _, log := range simResult.Value.Logs {
		fmt.Println("   ", log)
	}
}

// --- Реальна відправка (потребує funded акаунту) ---

func sendTransfer() {
	fmt.Println("\n=== Send SOL Transfer ===")

	// Замінити на реальний приватний ключ з балансом на devnet
	// Отримати SOL: solana airdrop 2 <PUBKEY> --url devnet
	sender := solana.NewWallet()
	receiver := solana.NewWallet()

	fmt.Printf("  Sender:   %s\n", sender.PublicKey())
	fmt.Printf("  Receiver: %s\n", receiver.PublicKey())

	// Перевіряємо баланс
	balOut, err := client.GetBalance(ctx, sender.PublicKey(), rpc.CommitmentConfirmed)
	if err != nil {
		fmt.Println("  GetBalance error:", err)
		return
	}
	fmt.Printf("  Sender balance: %d lamports\n", balOut.Value)

	if balOut.Value < solana.LAMPORTS_PER_SOL/10 {
		fmt.Println("  Недостатньо SOL. Запитай airdrop:")
		fmt.Printf("  solana airdrop 2 %s --url devnet\n", sender.PublicKey())
		return
	}

	// Blockhash
	bhOut, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("  GetLatestBlockhash error:", err)
		return
	}

	// Побудова
	amount := solana.LAMPORTS_PER_SOL / 100 // 0.01 SOL
	instruction := system.NewTransferInstruction(
		amount,
		sender.PublicKey(),
		receiver.PublicKey(),
	).Build()

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		bhOut.Value.Blockhash,
		solana.TransactionPayer(sender.PublicKey()),
	)
	if err != nil {
		fmt.Println("  build error:", err)
		return
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(sender.PublicKey()) {
			return &sender.PrivateKey
		}
		return nil
	})
	if err != nil {
		fmt.Println("  sign error:", err)
		return
	}

	// Відправка
	sig, err := client.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentProcessed,
	})
	if err != nil {
		fmt.Println("  send error:", err)
		return
	}
	fmt.Println("  Signature:", sig)
	fmt.Printf("  Explorer: https://explorer.solana.com/tx/%s?cluster=devnet\n", sig)

	// Polling підтвердження
	fmt.Println("  Чекаємо підтвердження...")
	if err := waitConfirmed(sig); err != nil {
		fmt.Println("  confirm error:", err)
		return
	}
	fmt.Println("  Підтверджено!")

	// Фінальний баланс
	newBal, _ := client.GetBalance(ctx, receiver.PublicKey(), rpc.CommitmentConfirmed)
	fmt.Printf("  Receiver balance: %d lamports\n", newBal.Value)
}

func waitConfirmed(sig solana.Signature) error {
	for range 30 {
		time.Sleep(2 * time.Second)
		out, err := client.GetSignatureStatuses(ctx, false, sig)
		if err != nil {
			return err
		}
		if out == nil || len(out.Value) == 0 || out.Value[0] == nil {
			continue
		}
		status := out.Value[0]
		if status.Err != nil {
			return fmt.Errorf("transaction failed: %v", status.Err)
		}
		if status.ConfirmationStatus == rpc.ConfirmationStatusConfirmed ||
			status.ConfirmationStatus == rpc.ConfirmationStatusFinalized {
			return nil
		}
	}
	return fmt.Errorf("timeout waiting for confirmation")
}
