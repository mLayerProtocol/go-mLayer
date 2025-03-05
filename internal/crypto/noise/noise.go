// File: noise/protocol.go
package noise

import (
	"crypto/rand"
	"errors"
	"io"

	"github.com/flynn/noise"
)

const (
    MaxMsgSize = 65535
)

// NoiseSession represents a noise protocol session using Flynn's implementation
type NoiseSession struct {
    hs   *noise.HandshakeState
    cs   *noise.CipherState
    keys noise.DHKey
}

// HybridEnvelope wraps a message with encryption for multiple recipients
type HybridEnvelope struct {
    Recipients [][]byte // Encrypted session keys for each recipient
    Payload    []byte   // Encrypted message payload
    Nonce      []byte   // Nonce used for payload encryption
}

// Initialize a new noise session
func NewNoiseSession() (*NoiseSession, error) {
    // Create new keypair
    dhKey, err := noise.DH25519.GenerateKeypair(rand.Reader)
    if err != nil {
        return nil, err
    }

    // Initialize Noise config
    config := noise.Config{
        CipherSuite: noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256),
        Pattern:     noise.HandshakeNN,
        Initiator:   true,
        StaticKeypair: dhKey,
    }

    // Create handshake state
    hs, err := noise.NewHandshakeState(config)
    if err != nil {
        return nil, err
    }

    return &NoiseSession{
        hs:   hs,
        keys: dhKey,
    }, nil
}

// EncryptMessage encrypts a message for multiple recipients
func (ns *NoiseSession) EncryptMessage(msg []byte, recipientKeys [][]byte) (*HybridEnvelope, error) {
    if len(msg) > MaxMsgSize {
        return nil, errors.New("message too large")
    }

    // Generate ephemeral key for this message
    messageKey := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, messageKey); err != nil {
        return nil, err
    }

    // Generate nonce
    nonce := make([]byte, 24)
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    // Create cipher for payload encryption
    cipher := noise.CipherSuite.Cipher(messageKey)
    payload := cipher.Encrypt(nil, nonce, nil, msg)

    // Encrypt message key for each recipient
    recipients := make([][]byte, len(recipientKeys))
    for i, recipientKey := range recipientKeys {
        encrypted, err := ns.encryptKeyForRecipient(messageKey, recipientKey)
        if err != nil {
            return nil, err
        }
        recipients[i] = encrypted
    }

    return &HybridEnvelope{
        Recipients: recipients,
        Payload:    payload,
        Nonce:      nonce,
    }, nil
}

// DecryptMessage decrypts a hybrid envelope message
func (ns *NoiseSession) DecryptMessage(env *HybridEnvelope) ([]byte, error) {
    // Try to decrypt the message key from any of the recipient slots
    var messageKey []byte
    var err error

    for _, recipientKey := range env.Recipients {
        messageKey, err = ns.decryptKeyForRecipient(recipientKey)
        if err == nil {
            break
        }
    }

    if messageKey == nil {
        return nil, errors.New("no valid recipient key found")
    }

    // Create cipher for payload decryption
    cipher := noise.CipherSuite.Cipher(messageKey)
    return cipher.Decrypt(nil, env.Nonce, nil, env.Payload)
}

// Helper functions for key encryption/decryption using noise protocol
func (ns *NoiseSession) encryptKeyForRecipient(messageKey, recipientKey []byte) ([]byte, error) {
    // Create ephemeral handshake state for key encryption
    config := noise.Config{
        CipherSuite: noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256),
        Pattern:     noise.HandshakeN,
        Initiator:   true,
        StaticKeypair: ns.keys,
    }

    hs, err := noise.NewHandshakeState(config)
    if err != nil {
        return nil, err
    }

    // Perform handshake and encrypt key
    msg, _, _, err := hs.WriteMessage(nil, messageKey)
    if err != nil {
        return nil, err
    }

    return msg, nil
}

func (ns *NoiseSession) decryptKeyForRecipient(encryptedKey []byte) ([]byte, error) {
    // Create handshake state for key decryption
    config := noise.Config{
        CipherSuite: noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256),
        Pattern:     noise.HandshakeN,
        Initiator:   false,
        StaticKeypair: ns.keys,
    }

    hs, err := noise.NewHandshakeState(config)
    if err != nil {
        return nil, err
    }

    // Perform handshake and decrypt key
    messageKey, _, _, err := hs.ReadMessage(nil, encryptedKey)
    if err != nil {
        return nil, err
    }

    return messageKey, nil
}