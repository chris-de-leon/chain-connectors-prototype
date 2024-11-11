package solana_testutils

import (
	"context"
	"errors"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
)

type AugmentedAccount struct {
	*Account
	*Backend
}

func (acct *AugmentedAccount) SignTx(tx *solana.Transaction) error {
	if _, err := tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if acct.PrivateKey.PublicKey().Equals(key) {
				return &acct.PrivateKey
			} else {
				return nil
			}
		},
	); err != nil {
		return err
	} else {
		return nil
	}
}

func (acct *AugmentedAccount) WaitForTxConfirm(ctx context.Context, sig solana.Signature) (solana.Signature, error) {
	confirmed, err := confirm.WaitForConfirmation(ctx, acct.WssClient, sig, &DefaultConfirmationTimeout)
	if err != nil {
		return sig, err
	}

	if !confirmed {
		return sig, errors.New("failed to confirm transaction")
	} else {
		return sig, nil
	}
}

func (acct *AugmentedAccount) SendAndConfirmTx(ctx context.Context, tx *solana.Transaction) (solana.Signature, error) {
	sig, err := confirm.SendAndConfirmTransaction(
		ctx,
		acct.RpcClient,
		acct.WssClient,
		tx,
	)
	if err != nil {
		return sig, err
	} else {
		return acct.WaitForTxConfirm(ctx, sig)
	}
}

func (acct *AugmentedAccount) SignAndSendTx(ctx context.Context, tx *solana.Transaction) (solana.Signature, error) {
	if err := acct.SignTx(tx); err != nil {
		return solana.Signature{}, err
	} else {
		return acct.SendAndConfirmTx(ctx, tx)
	}
}

func (acct *AugmentedAccount) FundAccount(ctx context.Context, sol uint64) (solana.Signature, error) {
	sig, err := acct.RpcClient.RequestAirdrop(ctx, acct.PrivateKey.PublicKey(), solana.LAMPORTS_PER_SOL*sol, rpc.CommitmentFinalized)
	if err != nil {
		return sig, err
	} else {
		return acct.WaitForTxConfirm(ctx, sig)
	}
}

func (acct *AugmentedAccount) TransferTokens(ctx context.Context, recipient solana.PublicKey, sol uint64) (solana.Signature, error) {
	pubKey := acct.PrivateKey.PublicKey()

	block, err := acct.RpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				solana.LAMPORTS_PER_SOL*sol,
				pubKey,
				recipient,
			).Build(),
		},
		block.Value.Blockhash,
		solana.TransactionPayer(pubKey),
	)
	if err != nil {
		return solana.Signature{}, err
	} else {
		return acct.SignAndSendTx(ctx, tx)
	}
}
