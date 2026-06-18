package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
)

type KeyPair struct {
	Public  [32]byte
	Private [32]byte
}

func GenerateKeyPair() (*KeyPair, error) {
	kp := &KeyPair{}
	if _, err := rand.Read(kp.Private[:]); err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}
	curve25519.ScalarBaseMult(&kp.Public, &kp.Private)
	return kp, nil
}

func SharedSecret(private [32]byte, public [32]byte) [32]byte {
	var secret [32]byte
	curve25519.ScalarMult(&secret, &private, &public)
	return secret
}

func DeriveKey(secret [32]byte) []byte {
	hash := sha256.Sum256(secret[:])
	return hash[:]
}

type EncryptedConn struct {
	rw       io.ReadWriter
	aead     cipher.AEAD
	nonceBuf []byte
}

func NewEncryptedConn(rw io.ReadWriter, key []byte) (*EncryptedConn, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new GCM: %w", err)
	}

	return &EncryptedConn{
		rw:       rw,
		aead:     aead,
		nonceBuf: make([]byte, aead.NonceSize()),
	}, nil
}

func (ec *EncryptedConn) Write(plaintext []byte) (int, error) {
	if _, err := rand.Read(ec.nonceBuf); err != nil {
		return 0, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := ec.aead.Seal(ec.nonceBuf[:0], ec.nonceBuf, plaintext, nil)
	
	lenBuf := make([]byte, 4)
	lenBuf[0] = byte(len(ciphertext) >> 24)
	lenBuf[1] = byte(len(ciphertext) >> 16)
	lenBuf[2] = byte(len(ciphertext) >> 8)
	lenBuf[3] = byte(len(ciphertext))

	if _, err := ec.rw.Write(lenBuf); err != nil {
		return 0, err
	}
	if _, err := ec.rw.Write(ciphertext); err != nil {
		return 0, err
	}

	return len(plaintext), nil
}

func (ec *EncryptedConn) Read(plaintext []byte) (int, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(ec.rw, lenBuf); err != nil {
		return 0, err
	}

	ciphertextLen := int(lenBuf[0])<<24 | int(lenBuf[1])<<16 | int(lenBuf[2])<<8 | int(lenBuf[3])
	ciphertext := make([]byte, ciphertextLen)
	if _, err := io.ReadFull(ec.rw, ciphertext); err != nil {
		return 0, err
	}

	nonceSize := ec.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return 0, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	decrypted, err := ec.aead.Open(plaintext[:0], nonce, ciphertext, nil)
	if err != nil {
		return 0, fmt.Errorf("decrypt: %w", err)
	}

	return len(decrypted), nil
}

func EncodePublicKey(pk [32]byte) string {
	return base64.RawURLEncoding.EncodeToString(pk[:])
}

func DecodePublicKey(s string) ([32]byte, error) {
	var pk [32]byte
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return pk, fmt.Errorf("decode key: %w", err)
	}
	if len(b) != 32 {
		return pk, fmt.Errorf("invalid key length: %d", len(b))
	}
	copy(pk[:], b)
	return pk, nil
}
