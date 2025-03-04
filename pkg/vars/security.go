package vars

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// SecurityManager 定义变量安全管理器
type SecurityManager struct {
	encryptionKey []byte
	sensitiveKeys []string
}

// NewSecurityManager 创建新的安全管理器
func NewSecurityManager(key string) *SecurityManager {
	// 使用SHA-256生成固定长度的密钥
	hash := sha256.Sum256([]byte(key))
	return &SecurityManager{
		encryptionKey: hash[:],
		sensitiveKeys: []string{"password", "secret", "key", "token", "credential"},
	}
}

// AddSensitiveKey 添加敏感变量关键字
func (sm *SecurityManager) AddSensitiveKey(key string) {
	for _, k := range sm.sensitiveKeys {
		if k == key {
			return
		}
	}
	sm.sensitiveKeys = append(sm.sensitiveKeys, key)
}

// IsSensitiveKey 检查是否为敏感变量
func (sm *SecurityManager) IsSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	for _, k := range sm.sensitiveKeys {
		if strings.Contains(key, k) {
			return true
		}
	}
	return false
}

// Encrypt 加密变量值
func (sm *SecurityManager) Encrypt(plaintext string) (string, error) {
	// 创建加密分组
	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 创建随机数
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密变量值
func (sm *SecurityManager) Decrypt(encrypted string) (string, error) {
	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	// 创建加密分组
	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 检查长度
	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("密文太短")
	}

	// 分离nonce和密文
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// SecureVariable 定义安全变量
type SecureVariable struct {
	BaseVariable
	encrypted bool
	security  *SecurityManager
}

// NewSecureVariable 创建新的安全变量
func NewSecureVariable(name string, value interface{}, varType VarType, security *SecurityManager) *SecureVariable {
	return &SecureVariable{
		BaseVariable: BaseVariable{
			Name:  name,
			Value: value,
			Type:  varType,
		},
		encrypted: security.IsSensitiveKey(name),
		security:  security,
	}
}

// GetValue 获取变量值（如果是敏感变量则解密）
func (sv *SecureVariable) GetValue() interface{} {
	if !sv.encrypted {
		return sv.Value
	}

	// 如果是字符串类型且已加密，则解密
	if sv.Type == VarTypeString {
		encrypted, ok := sv.Value.(string)
		if !ok {
			return sv.Value
		}

		// 检查是否已加密（以!ENCRYPTED:开头）
		if !strings.HasPrefix(encrypted, "!ENCRYPTED:") {
			return sv.Value
		}

		// 解密
		encrypted = strings.TrimPrefix(encrypted, "!ENCRYPTED:")
		decrypted, err := sv.security.Decrypt(encrypted)
		if err != nil {
			// 解密失败，返回原值
			return sv.Value
		}

		return decrypted
	}

	return sv.Value
}

// SetValue 设置变量值（如果是敏感变量则加密）
func (sv *SecureVariable) SetValue(value interface{}) error {
	// 如果不是敏感变量或不是字符串类型，直接设置
	if !sv.encrypted || sv.Type != VarTypeString {
		sv.Value = value
		return nil
	}

	// 如果是字符串类型的敏感变量，则加密
	strVal, ok := value.(string)
	if !ok {
		sv.Value = value
		return nil
	}

	// 检查是否已加密
	if strings.HasPrefix(strVal, "!ENCRYPTED:") {
		sv.Value = strVal
		return nil
	}

	// 加密
	encrypted, err := sv.security.Encrypt(strVal)
	if err != nil {
		return fmt.Errorf("加密变量失败: %w", err)
	}

	// 存储加密后的值
	sv.Value = "!ENCRYPTED:" + encrypted
	return nil
}