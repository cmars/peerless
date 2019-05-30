package peerless

import (
	"crypto/rand"
	"encoding/base32"

	"github.com/pkg/errors"
	"github.com/pquerna/otp/hotp"
)

type Authorization struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

func NewAuthorization() (*Authorization, error) {
	var token [32]byte
	_, err := rand.Reader.Read(token[:])
	if err != nil {
		return nil, errors.Wrap(err, "failed to read random bytes, heat death of universe imminent")
	}
	secret := base32.StdEncoding.EncodeToString(token[:])
	code, err := hotp.GenerateCode(secret, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate initial hotp code")
	}
	return &Authorization{Secret: secret, Code: code}, nil
}

func (a *Authorization) Next(counter uint64) error {
	code, err := hotp.GenerateCode(a.Secret, counter)
	if err != nil {
		return errors.Wrap(err, "failed to generate next hotp code")
	}
	a.Code = code
	return nil
}
