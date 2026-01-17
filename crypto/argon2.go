package crypto

import (
	"github.com/alexedwards/argon2id"
)

// PasswordManager 密码管理器
// Argon2 用于密码的哈希和验证

type PasswordManager struct {
	params *argon2id.Params
}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		params: &argon2id.Params{
			Memory:      64 * 1024, //生产环境使用
			Iterations:  3,
			Parallelism: 4,
			SaltLength:  16,
			KeyLength:   32,
		},
	}
}

// / Hash 对密码进行哈希
func (pm *PasswordManager) Hash(password string) (string, error) {
	return argon2id.CreateHash(password, pm.params)
}

// / Verify 验证密码是否匹配哈希值
func (pm *PasswordManager) Verify(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
