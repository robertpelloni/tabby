export enum SSHAlgorithmType {
    HMAC = 'hmac',
    KEX = 'kex',
    CIPHER = 'cipher',
    HOSTKEY = 'hostkey',
    COMPRESSION = 'compression',
}

// Stubs for the frontend dropdowns. The Go backend configures the real algorithms.
export const SupportedAlgorithms: Record<SSHAlgorithmType, string[]> = {
    [SSHAlgorithmType.KEX]: ['curve25519-sha256', 'ecdh-sha2-nistp256', 'ecdh-sha2-nistp384', 'ecdh-sha2-nistp521', 'diffie-hellman-group-exchange-sha256', 'diffie-hellman-group14-sha256'],
    [SSHAlgorithmType.HOSTKEY]: ['ssh-ed25519', 'ecdsa-sha2-nistp256', 'ecdsa-sha2-nistp384', 'ecdsa-sha2-nistp521', 'rsa-sha2-512', 'rsa-sha2-256', 'ssh-rsa'],
    [SSHAlgorithmType.CIPHER]: ['chacha20-poly1305@openssh.com', 'aes128-gcm@openssh.com', 'aes256-gcm@openssh.com', 'aes128-ctr', 'aes192-ctr', 'aes256-ctr'],
    [SSHAlgorithmType.HMAC]: ['hmac-sha2-256-etm@openssh.com', 'hmac-sha2-512-etm@openssh.com', 'hmac-sha2-256', 'hmac-sha2-512', 'hmac-sha1'],
    [SSHAlgorithmType.COMPRESSION]: ['none', 'zlib@openssh.com', 'zlib'],
}
