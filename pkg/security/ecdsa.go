package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"hash"
)

type EllipticCurve struct {
	curve      elliptic.Curve
	privateKey *ecdsa.PrivateKey
}

type IEllipticCurve interface {
	// GenerateKeys generates a public and private key pair.
	GenerateKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error)
	// EncodePrivate Marshals the PrivateKey to ASN.1() Format then to hexadecimal string
	EncodePrivate(privKey *ecdsa.PrivateKey) (string, error)
	// EncodePublic Marshals the publicKey to  ASN.1 Format then to hexadecimal string
	EncodePublic(pubKey *ecdsa.PublicKey) (string, error)
	// DecodePrivate decode the hexadecimal private key to  ASN.1 Format then to private key
	DecodePrivate(hexEncodedPriv string) (*ecdsa.PrivateKey, error)
	// DecodePublic decode the hexadecimal public key to  ASN.1 Format then to private key
	DecodePublic(hexEncodedPub string) (*ecdsa.PublicKey, error)
	// VerifySignature compare the signature of the data with the data itself to validate its integrity using the public key
	VerifySignature(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error)
	//CreateSignature creates a signature from the passed data using the private key
	//this signature passed between service alongside with the data itself
	//users would use this signature to validate the integrity of the passed data
	CreateSignature(data []byte, privKey *ecdsa.PrivateKey) ([]byte, error)
}

func NewEllipticCurve() IEllipticCurve {
	return &EllipticCurve{
		curve:      elliptic.P256(),
		privateKey: new(ecdsa.PrivateKey),
	}
}

func (ec *EllipticCurve) GenerateKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	var err error
	privKey, err := ecdsa.GenerateKey(ec.curve, rand.Reader)

	if err == nil {
		ec.privateKey = privKey
	}

	return ec.privateKey, &privKey.PublicKey, err
}

func (ec *EllipticCurve) EncodePrivate(privKey *ecdsa.PrivateKey) (string, error) {
	encoded, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encoded), nil
}

func (ec *EllipticCurve) EncodePublic(pubKey *ecdsa.PublicKey) (string, error) {
	encoded, err := x509.MarshalPKIXPublicKey(pubKey)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encoded), nil
}

func (ec *EllipticCurve) DecodePrivate(hexPrivateKey string) (*ecdsa.PrivateKey, error) {

	x509EncodedPriv, err := hex.DecodeString(hexPrivateKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := x509.ParseECPrivateKey(x509EncodedPriv)
	if err != nil {
		return nil, err
	}

	return privateKey, err
}

func (ec *EllipticCurve) DecodePublic(hexPublicKey string) (*ecdsa.PublicKey, error) {

	x509EncodedPub, err := hex.DecodeString(hexPublicKey)
	if err != nil {
		return nil, err
	}

	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return nil, err
	}

	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey, err
}

func (ec *EllipticCurve) CreateSignature(data []byte, privKey *ecdsa.PrivateKey) ([]byte, error) {

	var hasher hash.Hash
	hasher = sha256.New()

	hasher.Write(data)
	hashedData := hasher.Sum(nil)
	signature, err := ecdsa.SignASN1(rand.Reader, privKey, hashedData)

	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (ec *EllipticCurve) VerifySignature(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
	var hasher hash.Hash
	hasher = sha256.New()

	hasher.Write(data)
	hashedData := hasher.Sum(nil)
	verify := ecdsa.VerifyASN1(pubKey, hashedData, signature)
	return verify, nil
}
