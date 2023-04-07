package prove

import (
	"bytes"
	"fmt"
	"github.com/consensys/gnark/constraint"
	"math/big"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/circuit"
	"github.com/bnb-chain/zkbnb-crypto/circuit/types"
)

func LoadProvingKey(filepath string) (pks []groth16.ProvingKey, err error) {
	logx.Info("start reading proving key")
	return groth16.ReadSegmentProveKey(ecc.BN254, filepath)
}

func LoadVerifyingKey(filepath string) (verifyingKey groth16.VerifyingKey, err error) {
	verifyingKey = groth16.NewVerifyingKey(ecc.BN254)
	f, _ := os.Open(filepath + ".vk.save")
	_, err = verifyingKey.ReadFrom(f)
	if err != nil {
		return verifyingKey, fmt.Errorf("read file error")
	}
	f.Close()

	return verifyingKey, nil
}

func LoadR1CSLen(filename string) (nbConstraints int, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return -1, fmt.Errorf("read file error")
	}
	defer f.Close()

	var value int
	_, err = fmt.Fscanf(f, "%d", &value)
	if err != nil {
		return -1, err
	}

	return value, nil
}

func GenerateProof(
	r1cs constraint.ConstraintSystem,
	provingKey []groth16.ProvingKey,
	verifyingKey groth16.VerifyingKey,
	cBlock *circuit.Block,
	session string,
) (proof groth16.Proof, err error) {
	// verify CryptoBlock
	bN, err := circuit.ChooseBN(len(cBlock.Txs))
	if err != nil {
		return proof, err
	}
	blockWitness, err := circuit.SetBlockWitness(cBlock, bN)
	if err != nil {
		return proof, err
	}
	var verifyWitness circuit.BlockConstraints
	verifyWitness.OldStateRoot = cBlock.OldStateRoot
	verifyWitness.NewStateRoot = cBlock.NewStateRoot
	verifyWitness.BlockCommitment = cBlock.BlockCommitment
	witness, err := frontend.NewWitness(&blockWitness, ecc.BN254.ScalarField())
	if err != nil {
		return proof, err
	}

	vWitness, err := frontend.NewWitness(&verifyWitness, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return proof, err
	}
	// This is for test
	//witnessFullBytes, _ := witness.MarshalJSON()
	//err = os.WriteFile("witness_full.save", witnessFullBytes, 0666)
	//witnessPubBytes, _ := vWitness.MarshalJSON()
	//err = os.WriteFile("witness_pub.save", witnessPubBytes, 0666)
	proof, err = groth16.ProveRoll(r1cs, provingKey[0], provingKey[1], witness, session, backend.WithHints(types.PubDataToBytes))
	if err != nil {
		return proof, err
	}
	err = groth16.Verify(proof, verifyingKey, vWitness)
	if err != nil {
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
