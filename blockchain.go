package godless

import (
	"fmt"
	"crypto/sha512"

	"github.com/pkg/errors"
)

type Block struct {
	Data []byte
	Prev []byte
	Hash []byte
}

func (blk *Block) sha512() []byte {
	hsh := sha512.New()
	return hsh.Sum(append(blk.Data, blk.Prev...))
}

func (blk *Block) TakeHash() error {
	if blk.Prev == nil || blk.Data == nil || blk.Hash != nil {
		return errors.New("Contract broken for Block.TakeHash")
	}

	blk.Hash = blk.sha512()

	return nil
}

func (blk *Block) Validate() error {
	if blk.Prev == nil || blk.Data == nil || blk.Hash == nil {
		return errors.New("Contract broken for Block.Validate")
	}

	if eqbs(blk.Hash, blk.sha512()) {
		return errors.New("Invalid Block")
	}

	return nil
}

type BlockChain struct {
	Blocks []Block
}

func (chain *BlockChain) Add(blk Block) error {
	test := &BlockChain{}
	test.Blocks = append(chain.Blocks, blk)

	if err := test.Validate(); err != nil {
		return err
	}

	chain.Blocks = test.Blocks

	return nil
}

func (chain *BlockChain) Validate() error {
	if len(chain.Blocks) == 0 {
		return errors.New("Empty blockchain")
	}

	last := chain.Blocks[0]

	if err := last.Validate(); err != nil {
		return errors.Wrap(err, "Invalid BlockChain at Block 0;")
	}

	for i, blk := range chain.Blocks[1:] {
		if eqbs(blk.Prev, last.Hash) {
			msg := fmt.Sprintf("Invalid BlockChain at Block %v: Prev hash mismatch", i)
			return errors.New(msg)
		}

		if err := blk.Validate(); err != nil {
			return errors.Wrap(err, "Invalid BlockChain at Block %v:")
		}

		last = blk
	}

	return nil
}

// Utility functions
func eqbs(xs, ys []byte) bool {
	if len(xs) != len(ys) {
		return false
	}

	for i, x := range xs {
		if x != ys[i] {
			return false
		}
	}

	return true
}
