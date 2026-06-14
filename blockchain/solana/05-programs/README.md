# 05 — Programs & PDA

## Programs (смарт-контракти)

У Solana програми — це акаунти з `Executable: true`. Всі дані зберігаються **окремо** від програми, у data-акаунтах.

```
Program Account (executable)  — код
Data Account                   — стан (окремий акаунт, власником якого є програма)
```

## PDA — Program Derived Address

Детермінована адреса, що **не має приватного ключа** — підписується самою програмою.

```go
// FindProgramAddress: перебирає nonce від 255 до 0 поки не знайде точку поза кривою
pda, bump, err := solana.FindProgramAddress(
    [][]byte{
        []byte("seed-prefix"),
        owner.Bytes(),
        mint.Bytes(),
    },
    programID,
)
```

**Bump** — nonce (0-255), що зберігається в акаунті для верифікації.

## Стандартні PDA патерни

```go
// Metadata (Metaplex)
pda, _, _ = solana.FindProgramAddress([][]byte{
    []byte("metadata"),
    metaplexProgramID.Bytes(),
    mint.Bytes(),
}, metaplexProgramID)

// Edition
pda, _, _ = solana.FindProgramAddress([][]byte{
    []byte("metadata"),
    metaplexProgramID.Bytes(),
    mint.Bytes(),
    []byte("edition"),
}, metaplexProgramID)

// ATA (власна реалізація)
pda, _, _ = solana.FindProgramAddress([][]byte{
    owner.Bytes(),
    solana.TokenProgramID.Bytes(),
    mint.Bytes(),
}, solana.SPLAssociatedTokenAccountProgramID)
```

## Custom Instruction

```go
// Кожна програма визначає свій формат instruction data
// Зазвичай: перший byte/8 bytes — discriminator (ідентифікатор інструкції)

type MyInstruction struct {
    Discriminator [8]byte
    Amount        uint64
    Recipient     solana.PublicKey
}

data, _ := borsh.Serialize(MyInstruction{
    Discriminator: [8]byte{...},
    Amount:        1000,
    Recipient:     recipientPubkey,
})

instr := solana.NewInstruction(
    programID,
    solana.AccountMetaSlice{
        solana.Meta(signer).SIGNER().WRITE(),
        solana.Meta(pda).WRITE(),
    },
    data,
)
```

## Anchor discriminator

Anchor-програми використовують sha256 для discriminator:

```go
// discriminator = sha256("global:<instruction_name>")[:8]
h := sha256.Sum256([]byte("global:initialize"))
discriminator := h[:8]
```

## CPI — Cross-Program Invocation

Коли програма A викликає програму B зсередини транзакції. З Go — будуєш інструкції для зовнішньої програми і відправляєш у тій самій транзакції.

```
Tx
└── Instruction 1: твоя програма → CPI → SystemProgram::CreateAccount
└── Instruction 2: TokenProgram::InitializeMint
└── Instruction 3: AssociatedTokenAccount::Create
```
