package model

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestVendorJSONNeverContainsAPIKey 是 C1 的回归测试。
// /api/pricing 是无需认证的公开接口，此前直接序列化整个 Vendor 模型，
// 把每个上游供应商 API Key 的密文连同 base_url 一并泄露给匿名调用方。
func TestVendorJSONNeverContainsAPIKey(t *testing.T) {
	v := Vendor{
		ID:          1,
		Name:        "APIMart",
		Description: "test vendor",
		BaseURL:     "https://api.example.com",
		APIKey:      "super-secret-ciphertext-do-not-leak",
		Status:      1,
	}

	encoded, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	out := string(encoded)

	if strings.Contains(out, "super-secret-ciphertext-do-not-leak") {
		t.Errorf("Vendor 序列化结果泄露了 APIKey: %s", out)
	}
	if strings.Contains(out, "api_key") {
		t.Errorf("Vendor 序列化结果包含 api_key 字段: %s", out)
	}
}

// TestPublicVendorHasNoSecretFields 校验对外 DTO 的字段白名单
func TestPublicVendorHasNoSecretFields(t *testing.T) {
	pv := PublicVendor{
		ID:          1,
		Name:        "APIMart",
		Description: "test",
		Status:      1,
	}

	encoded, err := json.Marshal(pv)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	out := string(encoded)

	for _, forbidden := range []string{"api_key", "base_url", "secret", "token"} {
		if strings.Contains(out, forbidden) {
			t.Errorf("PublicVendor 不应包含 %q 字段: %s", forbidden, out)
		}
	}
}

// TestPublicVendorOmitsEmptyDescription 空描述不应产生噪音字段
func TestPublicVendorOmitsEmptyDescription(t *testing.T) {
	pv := PublicVendor{ID: 1, Name: "APIMart", Status: 1}
	encoded, _ := json.Marshal(pv)
	if strings.Contains(string(encoded), "description") {
		t.Errorf("空 description 应被省略: %s", encoded)
	}
}
