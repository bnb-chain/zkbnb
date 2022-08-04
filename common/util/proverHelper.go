package util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/plonk"
	"math/big"
	"os"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/consensys/gnark-crypto/ecc"
	kzg_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
)

const (
	COO_MODE = 1
	COM_MODE = 2
)

func LoadProvingKey(filepath string, srsfilepath string) (pk plonk.ProvingKey, err error) {
	fmt.Println("start reading proving key")
	pk = plonk.NewProvingKey(ecc.BN254)
	f, _ := os.Open(filepath)
	_, err = pk.ReadFrom(f)
	if err != nil {
		return pk, errors.New("read file error")
	}
	f, _ = os.Open(srsfilepath)
	var srs kzg_bn254.SRS
	_, err = srs.ReadFrom(f)
	pk.InitKZG(&srs)
	f.Close()

	return pk, nil
}

func LoadVerifyingKey(filepath string, srsfilepath string) (verifyingKey plonk.VerifyingKey, err error) {
	verifyingKey = plonk.NewVerifyingKey(ecc.BN254)
	f, _ := os.Open(filepath)
	_, err = verifyingKey.ReadFrom(f)
	if err != nil {
		return verifyingKey, errors.New("read file error")
	}
	f, _ = os.Open(srsfilepath)
	var srs kzg_bn254.SRS
	_, err = srs.ReadFrom(f)
	verifyingKey.InitKZG(&srs)
	f.Close()

	return verifyingKey, nil
}

func GenerateProof(
	r1cs frontend.CompiledConstraintSystem,
	provingKey plonk.ProvingKey,
	verifyingKey plonk.VerifyingKey,
	cBlock *cryptoBlock.Block,
) (proof plonk.Proof, err error) {
	// verify CryptoBlock
	blockWitness, err := cryptoBlock.SetBlockWitness(cBlock)
	if err != nil {
		logx.Errorf("[GenerateProof] unable to set block witness: %s", err.Error())
		return proof, err
	}
	var verifyWitness cryptoBlock.BlockConstraints
	verifyWitness.OldStateRoot = cBlock.OldStateRoot
	verifyWitness.NewStateRoot = cBlock.NewStateRoot
	verifyWitness.BlockCommitment = cBlock.BlockCommitment
	witness, err := frontend.NewWitness(&blockWitness, ecc.BN254)
	if err != nil {
		logx.Errorf("[GenerateProof] unable to generate new witness: %s", err.Error())
		return proof, err
	}
	vWitness, err := frontend.NewWitness(&verifyWitness, ecc.BN254, frontend.PublicOnly())
	if err != nil {
		logx.Errorf("[GenerateProof] unable to generate new witness: %s", err.Error())
		return proof, err
	}
	elapse := time.Now()
	logx.Info("start proving")
	proof, err = plonk.Prove(r1cs, provingKey, witness, backend.WithHints(std.Keccak256, std.ComputeSLp))
	if err != nil {
		logx.Errorf("[GenerateProof] unable to make a proof: %s", err.Error())
		return proof, err
	}
	fmt.Println("finish proving: ", time.Since(elapse))
	elapse = time.Now()
	logx.Info("start verifying")
	err = plonk.Verify(proof, verifyingKey, vWitness)
	if err != nil {
		logx.Errorf("[GenerateProof] invalid block proof: %s", err.Error())
		return proof, err
	}
	fmt.Println("finish verifying: ", time.Since(elapse))

	return proof, nil
}

func VerifyProof(
	proof plonk.Proof,
	vk plonk.VerifyingKey,
	cBlock *cryptoBlock.Block,
) error {
	// verify CryptoBlock
	blockWitness, err := cryptoBlock.SetBlockWitness(cBlock)
	if err != nil {
		logx.Errorf("[VerifyProof] unable to set block witness: %s", err.Error())
		return err
	}

	var verifyWitness cryptoBlock.BlockConstraints
	verifyWitness.OldStateRoot = cBlock.OldStateRoot
	verifyWitness.NewStateRoot = cBlock.NewStateRoot
	verifyWitness.BlockCommitment = cBlock.BlockCommitment
	_, err = frontend.NewWitness(&blockWitness, ecc.BN254)
	if err != nil {
		logx.Errorf("[VerifyProof] unable to generate new witness: %s", err.Error())
		return err
	}

	vWitness, err := frontend.NewWitness(&verifyWitness, ecc.BN254, frontend.PublicOnly())
	if err != nil {
		logx.Errorf("[VerifyProof] unable to generate new witness: %s", err.Error())
		return err
	}

	err = plonk.Verify(proof, vk, vWitness)
	if err != nil {
		logx.Errorf("[VerifyProof] invalid block proof: %s", err.Error())
		return err
	}
	return nil
}

type PlonkFormattedProof struct {
	WireCommitments               [3][2]*big.Int
	GrandProductCommitment        [2]*big.Int
	QuotientPolyCommitments       [3][2]*big.Int
	WireValuesAtZeta              [3]*big.Int
	GrandProductAtZetaOmega       *big.Int
	QuotientPolynomialAtZeta      *big.Int
	LinearizationPolynomialAtZeta *big.Int
	PermutationPolynomialsAtZeta  [2]*big.Int
	OpeningAtZetaProof            [2]*big.Int
	OpeningAtZetaOmegaProof       [2]*big.Int
}

func (p *PlonkFormattedProof) ConvertToArray(res *[]*big.Int) {
	for i := 0; i < 3; i++ {
		*res = append(*res, p.WireCommitments[i][:]...)
	}
	*res = append(*res, p.GrandProductCommitment[:]...)
	for i := 0; i < 3; i++ {
		*res = append(*res, p.QuotientPolyCommitments[i][:]...)
	}
	*res = append(*res, p.WireValuesAtZeta[:]...)
	*res = append(*res, p.GrandProductAtZetaOmega)
	*res = append(*res, p.QuotientPolynomialAtZeta)
	*res = append(*res, p.LinearizationPolynomialAtZeta)
	*res = append(*res, p.PermutationPolynomialsAtZeta[:]...)
	*res = append(*res, p.OpeningAtZetaProof[:]...)
	*res = append(*res, p.OpeningAtZetaOmegaProof[:]...)
}

func FormatPlonkProof(oProof plonk.Proof) (proof *PlonkFormattedProof, err error) {
	proof = new(PlonkFormattedProof)
	const fpSize = 32
	var buf bytes.Buffer
	_, err = oProof.WriteTo(&buf)

	if err != nil {
		logx.Errorf("unable to format plonk proof: %s", err.Error())
		return nil, err
	}
	proofBytes := buf.Bytes()
	index := 0
	var g1point bn254.G1Affine
	for i := 0; i < 3; i++ {
		g1point.SetBytes(proofBytes[fpSize*index : fpSize*(index+1)])
		uncompressed := g1point.RawBytes()
		for j := 0; j < 2; j++ {
			proof.WireCommitments[i][j] = new(big.Int).SetBytes(uncompressed[fpSize*j : fpSize*(j+1)])
		}
		index += 1
	}

	g1point.SetBytes(proofBytes[fpSize*index : fpSize*(index+1)])
	uncompressed := g1point.RawBytes()
	proof.GrandProductCommitment[0] = new(big.Int).SetBytes(uncompressed[0:fpSize])
	proof.GrandProductCommitment[1] = new(big.Int).SetBytes(uncompressed[fpSize : fpSize*2])
	index += 1

	for i := 0; i < 3; i++ {
		g1point.SetBytes(proofBytes[fpSize*index : fpSize*(index+1)])
		uncompressed := g1point.RawBytes()
		for j := 0; j < 2; j++ {
			proof.QuotientPolyCommitments[i][j] = new(big.Int).SetBytes(uncompressed[fpSize*j : fpSize*(j+1)])
		}
		index += 1
	}

	g1point.SetBytes(proofBytes[fpSize*index : fpSize*(index+1)])
	uncompressed = g1point.RawBytes()
	proof.OpeningAtZetaProof[0] = new(big.Int).SetBytes(uncompressed[0:fpSize])
	proof.OpeningAtZetaProof[1] = new(big.Int).SetBytes(uncompressed[fpSize : fpSize*2])
	index += 1
	fmt.Printf("OpeningAtZetaProof is %x\n", uncompressed)

	// plonk proof write len(ClaimedValues) which is 4 bytes
	offset := 4
	fmt.Printf("QuotientPolynomialAtZeta is %x\n", proofBytes[offset+fpSize*index:offset+fpSize*(index+1)])
	proof.QuotientPolynomialAtZeta = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
	index += 1

	proof.LinearizationPolynomialAtZeta = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
	index += 1

	for i := 0; i < 3; i++ {
		proof.WireValuesAtZeta[i] = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
		index += 1
	}

	for i := 0; i < 2; i++ {
		proof.PermutationPolynomialsAtZeta[i] = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
		index += 1
	}

	g1point.SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
	uncompressed = g1point.RawBytes()
	proof.OpeningAtZetaOmegaProof[0] = new(big.Int).SetBytes(uncompressed[0:fpSize])
	proof.OpeningAtZetaOmegaProof[1] = new(big.Int).SetBytes(uncompressed[fpSize : fpSize*2])
	index += 1

	proof.GrandProductAtZetaOmega = new(big.Int).SetBytes(proofBytes[offset+fpSize*index : offset+fpSize*(index+1)])
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
		logx.Errorf("[FormatProof] unable to format proof: %s", err.Error())
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

func UnformatProof(proof *FormattedProof) (oProof plonk.Proof, err error) {
	var buf bytes.Buffer
	// write bytes to buffer
	buf.Write(proof.A[0].Bytes())
	buf.Write(proof.A[1].Bytes())
	buf.Write(proof.B[0][0].Bytes())
	buf.Write(proof.B[0][1].Bytes())
	buf.Write(proof.B[1][0].Bytes())
	buf.Write(proof.B[1][1].Bytes())
	buf.Write(proof.C[0].Bytes())
	buf.Write(proof.C[1].Bytes())

	// init oProof
	oProof = plonk.NewProof(ecc.BN254)

	// read buffer
	_, err = oProof.ReadFrom(bytes.NewReader(buf.Bytes()))
	if err != nil {
		logx.Errorf("[UnformatProof] unable to ReadFrom proof buffer: %s", err.Error())
		return oProof, err
	}

	return oProof, nil
}

func CompactProofs(proofs []*FormattedProof) []*big.Int {
	var res []*big.Int
	for _, proof := range proofs {
		res = append(res, proof.A[0])
		res = append(res, proof.A[1])
		res = append(res, proof.B[0][0])
		res = append(res, proof.B[0][1])
		res = append(res, proof.B[1][0])
		res = append(res, proof.B[1][1])
		res = append(res, proof.C[0])
		res = append(res, proof.C[1])
	}
	return res
}
