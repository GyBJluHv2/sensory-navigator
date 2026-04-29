package auth

import (
	"testing"
	"time"

	"github.com/atakhanov/sensory-navigator/backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestPasswordHashing(t *testing.T) {
	hash, err := GeneratePasswordHash("strong-pw", 4)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, CheckPassword("strong-pw", hash))
	assert.False(t, CheckPassword("wrong", hash))
}

func TestJWTRoundtrip(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", JWTAccessTTL: time.Minute}
	token, exp, err := IssueToken(42, cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, exp.After(time.Now()))

	claims, err := ParseToken(token, cfg)
	assert.NoError(t, err)
	assert.Equal(t, uint64(42), claims.UserID)
}

func TestJWTRejectsWrongSecret(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", JWTAccessTTL: time.Minute}
	token, _, _ := IssueToken(1, cfg)

	wrong := &config.Config{JWTSecret: "another", JWTAccessTTL: time.Minute}
	_, err := ParseToken(token, wrong)
	assert.ErrorIs(t, err, ErrInvalidToken)
}
