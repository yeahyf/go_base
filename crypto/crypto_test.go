package crypto

import (
	"strings"
	"testing"
)

func TestNewPasswordManager(t *testing.T) {
	pm := NewPasswordManager()
	if pm == nil {
		t.Fatal("NewPasswordManager returned nil")
	}
	if pm.params == nil {
		t.Fatal("params is nil")
	}
	if pm.params.Memory != 64*1024 {
		t.Errorf("expected Memory %d, got %d", 64*1024, pm.params.Memory)
	}
	if pm.params.Iterations != 3 {
		t.Errorf("expected Iterations %d, got %d", 3, pm.params.Iterations)
	}
	if pm.params.Parallelism != 4 {
		t.Errorf("expected Parallelism %d, got %d", 4, pm.params.Parallelism)
	}
	if pm.params.SaltLength != 16 {
		t.Errorf("expected SaltLength %d, got %d", 16, pm.params.SaltLength)
	}
	if pm.params.KeyLength != 32 {
		t.Errorf("expected KeyLength %d, got %d", 32, pm.params.KeyLength)
	}
}

func TestPasswordManager_Hash(t *testing.T) {
	pm := NewPasswordManager()
	password := "testPassword123!"

	hash, err := pm.Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	if hash == "" {
		t.Fatal("Hash returned empty string")
	}
	if !strings.Contains(hash, "$argon2id$") {
		t.Error("Hash does not contain argon2id identifier")
	}
}

func TestPasswordManager_Hash_EmptyPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := ""

	hash, err := pm.Hash(password)
	if err != nil {
		t.Fatalf("Hash with empty password failed: %v", err)
	}
	if hash == "" {
		t.Fatal("Hash returned empty string for empty password")
	}
}

func TestPasswordManager_Verify_CorrectPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := "correctPassword123!"

	hash, err := pm.Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	valid, err := pm.Verify(password, hash)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid {
		t.Error("Verify returned false for correct password")
	}
}

func TestPasswordManager_Verify_WrongPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := "correctPassword123!"
	wrongPassword := "wrongPassword456!"

	hash, err := pm.Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	valid, err := pm.Verify(wrongPassword, hash)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if valid {
		t.Error("Verify returned true for wrong password")
	}
}

func TestPasswordManager_Verify_EmptyPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := "testPassword123!"

	hash, err := pm.Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	valid, err := pm.Verify("", hash)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if valid {
		t.Error("Verify returned true for empty password")
	}
}

func TestPasswordManager_Verify_InvalidHash(t *testing.T) {
	pm := NewPasswordManager()
	password := "testPassword123!"
	invalidHash := "invalid_hash_format"

	_, err := pm.Verify(password, invalidHash)
	if err == nil {
		t.Error("Verify did not return error for invalid hash")
	}
}

func TestPasswordManager_HashAndVerify_DifferentPasswords(t *testing.T) {
	pm := NewPasswordManager()
	password1 := "password1"
	password2 := "password2"

	hash1, err := pm.Hash(password1)
	if err != nil {
		t.Fatalf("Hash failed for password1: %v", err)
	}

	hash2, err := pm.Hash(password2)
	if err != nil {
		t.Fatalf("Hash failed for password2: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Different passwords produced the same hash")
	}

	valid1, err := pm.Verify(password1, hash1)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid1 {
		t.Error("Verify failed for password1 with hash1")
	}

	valid2, err := pm.Verify(password2, hash2)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid2 {
		t.Error("Verify failed for password2 with hash2")
	}

	validCross1, err := pm.Verify(password1, hash2)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if validCross1 {
		t.Error("Verify returned true for password1 with hash2")
	}

	validCross2, err := pm.Verify(password2, hash1)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if validCross2 {
		t.Error("Verify returned true for password2 with hash1")
	}
}

func TestPasswordManager_Hash_SamePasswordDifferentHashes(t *testing.T) {
	pm := NewPasswordManager()
	password := "samePassword123!"

	hash1, err := pm.Hash(password)
	if err != nil {
		t.Fatalf("First Hash failed: %v", err)
	}

	hash2, err := pm.Hash(password)
	if err != nil {
		t.Fatalf("Second Hash failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Same password produced identical hashes (salt should make them different)")
	}

	valid1, err := pm.Verify(password, hash1)
	if err != nil {
		t.Fatalf("Verify failed for hash1: %v", err)
	}
	if !valid1 {
		t.Error("Verify failed for hash1")
	}

	valid2, err := pm.Verify(password, hash2)
	if err != nil {
		t.Fatalf("Verify failed for hash2: %v", err)
	}
	if !valid2 {
		t.Error("Verify failed for hash2")
	}
}

func TestPasswordManager_Hash_SpecialCharacters(t *testing.T) {
	pm := NewPasswordManager()
	passwords := []string{
		"密码123",
		"p@ssw0rd!#$%",
		"unicode🔐",
		"spaces in password",
		"tab\tcharacter",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hash, err := pm.Hash(password)
			if err != nil {
				t.Fatalf("Hash failed: %v", err)
			}

			valid, err := pm.Verify(password, hash)
			if err != nil {
				t.Fatalf("Verify failed: %v", err)
			}
			if !valid {
				t.Error("Verify returned false for correct password with special characters")
			}
		})
	}
}

func TestPasswordManager_Hash_LongPassword(t *testing.T) {
	pm := NewPasswordManager()
	longPassword := strings.Repeat("a", 1000)

	hash, err := pm.Hash(longPassword)
	if err != nil {
		t.Fatalf("Hash failed for long password: %v", err)
	}

	valid, err := pm.Verify(longPassword, hash)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid {
		t.Error("Verify returned false for long password")
	}
}

func BenchmarkPasswordManager_Hash(b *testing.B) {
	pm := NewPasswordManager()
	password := "testPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordManager_Verify(b *testing.B) {
	pm := NewPasswordManager()
	password := "testPassword123!"
	hash, err := pm.Hash(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.Verify(password, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordManager_HashAndVerify(b *testing.B) {
	pm := NewPasswordManager()
	password := "testPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash, err := pm.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
		_, err = pm.Verify(password, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordManager_Hash_ShortPassword(b *testing.B) {
	pm := NewPasswordManager()
	password := "short"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordManager_Hash_MediumPassword(b *testing.B) {
	pm := NewPasswordManager()
	password := "mediumPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordManager_Hash_LongPassword(b *testing.B) {
	pm := NewPasswordManager()
	password := strings.Repeat("a", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordManager_Verify_CorrectPassword(b *testing.B) {
	pm := NewPasswordManager()
	password := "testPassword123!"
	hash, err := pm.Hash(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.Verify(password, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordManager_Verify_WrongPassword(b *testing.B) {
	pm := NewPasswordManager()
	password := "testPassword123!"
	wrongPassword := "wrongPassword456!"
	hash, err := pm.Hash(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.Verify(wrongPassword, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordManager_Hash_Parallel(b *testing.B) {
	pm := NewPasswordManager()
	password := "testPassword123!"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := pm.Hash(password)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPasswordManager_Verify_Parallel(b *testing.B) {
	pm := NewPasswordManager()
	password := "testPassword123!"
	hash, err := pm.Hash(password)
	if err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := pm.Verify(password, hash)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPasswordManager_NewPasswordManager(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPasswordManager()
	}
}
