package protocol

import "crypto/cipher"

// cfb8 implements AES with 8-bit cipher feedback — Java's "AES/CFB8", the
// stream cipher Minecraft uses for login encryption. Go's crypto/cipher only
// ships full-block CFB, so this is the byte-feedback variant: each input
// byte is XORed with the first byte of E(shift register), and one byte is
// fed back into the register (ciphertext when decrypting, ciphertext output
// when encrypting — per CFB semantics the ciphertext always feeds back).
type cfb8 struct {
	b       cipher.Block
	iv      [16]byte
	out     [16]byte
	decrypt bool
}

// XORKeyStream implements cipher.Stream. dst and src may alias.
func (s *cfb8) XORKeyStream(dst, src []byte) {
	for i := range src {
		s.b.Encrypt(s.out[:], s.iv[:])
		c := src[i] ^ s.out[0]
		copy(s.iv[:], s.iv[1:])
		if s.decrypt {
			s.iv[15] = src[i]
		} else {
			s.iv[15] = c
		}
		dst[i] = c
	}
}
