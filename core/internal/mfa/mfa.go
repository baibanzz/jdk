package mfa

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/baibanzz/jdk/utils"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPEncryptAES 使用历史兼容的 AES 方式加密 TOTP secret。
//
// 加密 key 由 username/userId 左侧补空格到至少 16 位得到，并在密文前追加
// 与 username/userId 等长的随机字符串作为前缀；TOTPDecryptAES 会按相同长度
// 去掉前缀后再解密。这里必须保持旧算法不变，否则库中已绑定的 MFA 密钥无法解密。
//
// 当 AES key 长度不满足 16/24/32 字节时，底层加密会失败，本方法保持历史行为返回空字符串。
func TOTPEncryptAES(secret, username string) string {
	aesKey := fmt.Sprintf("%016s", username)
	encrypted, err := utils.EncryptAES([]byte(secret), []byte(aesKey))
	if err != nil {
		return ""
	}
	randStr := strings.ToUpper(utils.RandString(len(username)))
	return randStr + base64.StdEncoding.EncodeToString(encrypted)
}

// TOTPDecryptAES 解密由 TOTPEncryptAES 生成的 TOTP secret。
//
// data 前 len(username/userId) 位是随机前缀，剩余部分是 base64 编码后的 AES 密文。
// 解密失败、密文格式错误或前缀长度异常时返回空字符串，由业务层决定错误文案与处理方式。
func TOTPDecryptAES(data, username string) string {
	if len(data) < len(username) {
		return ""
	}
	secret := data[len(username):]
	secretByte, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return ""
	}
	aesKey := fmt.Sprintf("%016s", username)
	decrypted, err := utils.DecryptAES(secretByte, []byte(aesKey))
	if err != nil {
		return ""
	}
	return string(decrypted)
}

// GenerateTOTP 创建 Google Authenticator 兼容的 TOTP secret 与 otpauth URL。
//
// issuer 会显示为验证器中的发行方，account 会显示为账号名；返回的 qrURL 可用于生成二维码。
// 本方法只负责生成密钥，不负责保存、绑定或权限校验。
func GenerateTOTP(issuer, account string) (secret, qrURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// VerifyTOTP 校验后台 admin 侧 6 位 TOTP 验证码。
//
// 保留原后台 MFA 行为：30 秒周期、SHA1、6 位数字，并允许前后 2 个时间窗口的偏移。
// 业务层负责先解密 secret、校验验证码格式、限制错误次数等流程控制。
func VerifyTOTP(secret, code string) (bool, error) {
	return verifyTOTP(secret, code, 2)
}

// VerifyUserTOTP 校验前台用户侧 6 位 TOTP 验证码。
//
// 保留原用户 MFA 行为：30 秒周期、SHA1、6 位数字，并允许前后 1 个时间窗口的偏移。
// 这个窗口比后台 admin 侧更小，避免迁移后改变用户侧历史校验语义。
func VerifyUserTOTP(secret, code string) (bool, error) {
	return verifyTOTP(secret, code, 1)
}

// verifyTOTP 按指定时间偏移窗口校验 TOTP 验证码。
func verifyTOTP(secret, code string, skew uint) (bool, error) {
	return totp.ValidateCustom(code, secret, time.Now(), validateOpts(skew))
}

// validateOpts 返回项目历史固定的 TOTP 参数。
func validateOpts(skew uint) totp.ValidateOpts {
	return totp.ValidateOpts{
		Period:    30,
		Skew:      skew,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
}
