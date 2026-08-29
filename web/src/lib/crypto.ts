/**
 * AES-GCM 解密（与后端 common.EncryptWithKey 对应）
 * @param encryptedBase64 - base64 编码的密文（nonce + ciphertext）
 * @param keyBase64 - base64 编码的 32 字节密钥
 */
export async function decryptWithKey(encryptedBase64: string, keyBase64: string): Promise<string> {
  const keyBytes = Uint8Array.from(atob(keyBase64), (c) => c.charCodeAt(0))
  const encryptedBytes = Uint8Array.from(atob(encryptedBase64), (c) => c.charCodeAt(0))

  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    keyBytes,
    { name: 'AES-GCM' },
    false,
    ['decrypt']
  )

  // AES-GCM: 前 12 字节是 nonce
  const nonce = encryptedBytes.slice(0, 12)
  const ciphertext = encryptedBytes.slice(12)

  const plaintext = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: nonce },
    cryptoKey,
    ciphertext
  )

  return new TextDecoder().decode(plaintext)
}
