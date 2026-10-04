import JSEncrypt from 'jsencrypt'

// RSA 公钥（与后端 configs/*.yaml 中的 rsa.public_key 保持一致）
const publicKey =
  'MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1gSEu/3HAVyof9+SAhBa' +
  'T3qKoRAqrofLzgU4hiDxMDKi9zxF7shCMaQE1a62zdDMHJna3I9N6uqZeENtVRoE' +
  'lJvbOEvCmtRU1GYV51AztmCBFOrpP4Qlg2o9l2eZ1D9C4k/MjFfulg+AF6DYVKjA' +
  'xvX+da2iImrVhs8h453MTcFIL/OU9LBQPt81Ii7ynqKsjFIKHuZhQ6e7y+q25UKS' +
  'h/dg41rsUrbsSVmndg5cHfBVIsMJVQg5/dt5td82E4NurZe1iMWsBYYHG5b152AV' +
  'dD5j0xa08qXx5H3p/777Klp2vNhEl+b2mK0rJTdf4GAOdfRaoQIGL/mlqGZ+Vndl' +
  'hwIDAQAB'

/**
 * RSA 加密
 * @param text 明文
 * @returns Base64 编码的密文
 */
export function encrypt(text: string): string {
  const encryptor = new JSEncrypt()
  encryptor.setPublicKey(publicKey)
  const result = encryptor.encrypt(text)
  return result === false ? '' : result
}
