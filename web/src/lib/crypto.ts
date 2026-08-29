/**
 * AES-GCM 加密（与后端 common.DecryptWithKey 对应）
 * @param plaintext - 明文
 * @param keyBase64 - base64 编码的 32 字节密钥
 */
export async function encryptWithKey(plaintext: string, keyBase64: string): Promise<string> {
  const keyBytes = Uint8Array.from(atob(keyBase64), (c) => c.charCodeAt(0))

  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    keyBytes,
    { name: 'AES-GCM' },
    false,
    ['encrypt']
  )

  const nonce = crypto.getRandomValues(new Uint8Array(12))
  const encoded = new TextEncoder().encode(plaintext)

  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv: nonce },
    cryptoKey,
    encoded
  )

  // nonce + ciphertext 拼接后 base64
  const combined = new Uint8Array(nonce.length + ciphertext.byteLength)
  combined.set(nonce)
  combined.set(new Uint8Array(ciphertext), nonce.length)

  return btoa(String.fromCharCode(...combined))
}

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
