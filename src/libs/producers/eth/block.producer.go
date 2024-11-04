package eth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"google.golang.org/grpc"
)

type BlockProducer struct {
	proto.UnimplementedBlockProducerServer
	stream *BlockStreamer
}

func NewBlockProducer(stream *BlockStreamer) BlockProducer {
	return BlockProducer{
		stream: stream,
	}
}

func (producer BlockProducer) RegisterToServer(srv *grpc.Server) *grpc.Server {
	proto.RegisterBlockProducerServer(srv, producer)
	return srv
}

func (producer BlockProducer) Blocks(initBlock *proto.InitBlock, stream grpc.ServerStreamingServer[proto.Block]) error {
	ctx := stream.Context()

	cursor, err := producer.initCursor(ctx, initBlock)
	if err != nil {
		return err
	}

	for {
		block, err := producer.stream.WaitForNextBlock(ctx, cursor)
		if err != nil {
			return err
		}

		data, err := producer.stringifyBlock(block)
		if err != nil {
			return err
		}

		err = stream.Send(&proto.Block{
			Height: block.Number().String(),
			Data:   string(data),
		})
		if err != nil {
			return err
		}

		cursor = new(big.Int).Add(cursor, big.NewInt(1))
	}
}

func (producer BlockProducer) initCursor(ctx context.Context, initBlock *proto.InitBlock) (*big.Int, error) {
	latestBlock, err := producer.stream.client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	if latestBlock == nil {
		return nil, errors.New("failed to retrieve the latest block")
	}
	if initBlock.Height == nil {
		return latestBlock.Number(), nil
	}

	cursor, ok := new(big.Int).SetString(initBlock.GetHeight(), 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert string '%s' to big int", initBlock.GetHeight())
	}
	if cursor.Cmp(big.NewInt(0)) == -1 {
		return nil, errors.New("invalid start block height (must be >= 0)")
	}
	if cursor.Cmp(latestBlock.Number()) == 1 {
		return nil, errors.New("starting block height is larger than latest block height")
	}

	return cursor, nil
}

func (producer BlockProducer) stringifyBlock(block *ethtypes.Block) ([]byte, error) {
	return json.Marshal(map[string]any{
		"receivedAt": block.ReceivedAt.String(),
		"baseFee":    block.BaseFee().String(),
		// "beaconRoot":     block.BeaconRoot().String(), // throws an error
		"blobGasUsed": block.BlobGasUsed(),
		"bloom":       block.Bloom().Bytes(),
		// "body":            block.Body(), // redundant
		"coinbase":      block.Coinbase().String(),
		"difficulty":    block.Difficulty().String(),
		"excessBlobGas": block.ExcessBlobGas(),
		"extra":         block.Extra(),
		"gasLimit":      block.GasLimit(),
		"gasUsed":       block.GasUsed(),
		"hash":          block.Hash().String(),
		"header":        block.Header(),
		"mixDigest":     block.MixDigest().String(),
		"nonce":         block.Nonce(),
		"number":        block.Number(),
		"parentHash":    block.ParentHash().String(),
		"receiptHash":   block.ReceiptHash().String(),
		"root":          block.Root().String(),
		"size":          block.Size(),
		"time":          block.Time(),
		"transactions":  block.Transactions(),
		"txHash":        block.TxHash().String(),
		"uncleHash":     block.UncleHash().String(),
		"uncles":        block.Uncles(),
		"withdrawals":   block.Withdrawals(),
	})
}
