package core

import (
	"github.com/baibanzz/jdk/core/internal/mfa"
)

var Name string

type MFA struct {
	Secret  string //MFA原始秘钥
	User    string //用户名
	SaveKey string //私钥
}

func NewMFA(user string) *MFA {
	return &MFA{User: user}
}

// Create 生成私钥
func (m *MFA) Create() (priKey string, urlPath string, err error) {
	secret, url, err := mfa.GenerateTOTP(Name, m.User)
	m.Secret = secret
	return secret, url, err
}

// getSecretBySaveKey 通过存储的私钥+用户名获取原始秘钥
func (m *MFA) getSecretBySaveKey() string {
	return mfa.TOTPDecryptAES(m.SaveKey, m.User)
}

// getSaveKeyBySecret 通过原始秘钥+用户名获取存储的私钥
func (m *MFA) getSaveKeyBySecret() string {
	return mfa.TOTPEncryptAES(m.Secret, m.User)
}

// Auto 通过已有数据补全其它数据
func (m *MFA) Auto() *MFA {
	if m.User != "" {
		if m.Secret == "" && m.SaveKey != "" {
			m.Secret = m.getSaveKeyBySecret()
			if m.Secret == "" {
				return nil
			}
		}
		if m.Secret != "" && m.SaveKey == "" {
			m.SaveKey = m.getSecretBySaveKey()
			if m.SaveKey == "" {
				return nil
			}
		}
		if m.SaveKey == "" && m.Secret == "" {
			return nil
		}
	} else {
		return nil
	}
	return m
}

// VerifyMFA 验证
func (m *MFA) VerifyMFA(code string) bool {
	ok, _ := mfa.VerifyTOTP(m.Secret, code)
	return ok
}

// MFAAfterFind 查询后自动解密 SaveKey → Secret
func MFAAfterFind(user, saveKey string) *MFA {
	mfa := NewMFA(user)
	mfa.SaveKey = saveKey
	return mfa.Auto()
}
