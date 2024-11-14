package jwtutils

import (
	"testing"
	"user_mgt/user_mgt/jwtutils"
)

func TestGenerateToken(t *testing.T, regid string) string {
	jwtServer := jwtutils.NewJWTserve()

	regToken, err := jwtServer.GenerateToken(regid, jwtutils.TokenTypeRegistered)
	if err != nil {
		t.Errorf("failed to generate token: %s", err)
	}
	t.Logf("reg token: %s", regToken)
	_, err = jwtServer.ParseAndVerifyToken(regToken, jwtutils.TokenTypeRegistered)
	if err != nil {
		t.Errorf("failed to parse and verify reg token: %s", err)
	}
	t.Logf("reg token is valid")

	return regToken
}
