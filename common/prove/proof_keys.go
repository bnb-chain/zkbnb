package prove

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/zeromicro/go-zero/core/logx"

	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
)

func LoadProvingKey(filepath string) (pk groth16.ProvingKey, err error) {
	fmt.Println("start reading proving key")
	pk = groth16.NewProvingKey(ecc.BN254)
	f, _ := os.Open(filepath)
	_, err = pk.ReadFrom(f)
	if err != nil {
		return pk, errors.New("read file error")
	}
	f.Close()

	return pk, nil
}

func LoadVerifyingKey(filepath string) (verifyingKey groth16.VerifyingKey, err error) {
	verifyingKey = groth16.NewVerifyingKey(ecc.BN254)
	f, _ := os.Open(filepath)
	_, err = verifyingKey.ReadFrom(f)
	if err != nil {
		return verifyingKey, errors.New("read file error")
	}
	f.Close()

	return verifyingKey, nil
}

func GenerateProof(
	r1cs frontend.CompiledConstraintSystem,
	provingKey groth16.ProvingKey,
	verifyingKey groth16.VerifyingKey,
	cBlock *cryptoBlock.Block,
) (proof groth16.Proof, err error) {
	// verify CryptoBlock
	blockWitness, err := cryptoBlock.SetBlockWitness(cBlock)
	if err != nil {
		logx.Errorf("unable to set block witness: %s", err.Error())
		return proof, err
	}
	var verifyWitness cryptoBlock.BlockConstraints
	verifyWitness.OldStateRoot = cBlock.OldStateRoot
	verifyWitness.NewStateRoot = cBlock.NewStateRoot
	verifyWitness.BlockCommitment = cBlock.BlockCommitment
	witness, err := frontend.NewWitness(&blockWitness, ecc.BN254)
	if err != nil {
		logx.Errorf("unable to generate new witness: %s", err.Error())
		return proof, err
	}
	vWitness, err := frontend.NewWitness(&verifyWitness, ecc.BN254, frontend.PublicOnly())
	if err != nil {
		logx.Errorf("unable to generate new witness: %s", err.Error())
		return proof, err
	}
	elapse := time.Now()
	logx.Info("start proving")
	proof, err = groth16.Prove(r1cs, provingKey, witness, backend.WithHints(std.Keccak256, std.ComputeSLp))
	if err != nil {
		logx.Errorf("unable to make a proof: %s", err.Error())
		return proof, err
	}
	fmt.Println("finish proving: ", time.Since(elapse))
	elapse = time.Now()
	logx.Info("start verifying")
	err = groth16.Verify(proof, verifyingKey, vWitness)
	if err != nil {
		logx.Errorf("invalid block proof: %s", err.Error())
		return proof, err
	}

	return proof, nil
}

type FormattedProof struct {
	A      [2]*big.Int
	B      [2][2]*big.Int
	C      [2]*big.Int
	Inputs [3]*big.Int
}

func FormatProof(oProof groth16.Proof, oldRoot, newRoot, commitment []byte) (proof *FormattedProof, err error) {
	proof = new(FormattedProof)
	const fpSize = 4 * 8
	var buf bytes.Buffer
	_, err = oProof.WriteRawTo(&buf)
	if err != nil {
		logx.Errorf("unable to format proof: %s", err.Error())
		return nil, err
	}
	proofBytes := buf.Bytes()
	// proof.Ar, proof.Bs, proof.Krs
	proof.A[0] = new(big.Int).SetBytes(proofBytes[fpSize*0 : fpSize*1])
	proof.A[1] = new(big.Int).SetBytes(proofBytes[fpSize*1 : fpSize*2])
	proof.B[0][0] = new(big.Int).SetBytes(proofBytes[fpSize*2 : fpSize*3])
	proof.B[0][1] = new(big.Int).SetBytes(proofBytes[fpSize*3 : fpSize*4])
	proof.B[1][0] = new(big.Int).SetBytes(proofBytes[fpSize*4 : fpSize*5])
	proof.B[1][1] = new(big.Int).SetBytes(proofBytes[fpSize*5 : fpSize*6])
	proof.C[0] = new(big.Int).SetBytes(proofBytes[fpSize*6 : fpSize*7])
	proof.C[1] = new(big.Int).SetBytes(proofBytes[fpSize*7 : fpSize*8])

	// public witness
	proof.Inputs[0] = new(big.Int).SetBytes(oldRoot)
	proof.Inputs[1] = new(big.Int).SetBytes(newRoot)
	proof.Inputs[2] = new(big.Int).SetBytes(commitment)
	return proof, nil
}
