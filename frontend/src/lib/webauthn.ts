function base64UrlToBuffer(value: string): ArrayBuffer {
  const base64 = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64.padEnd(
    base64.length + ((4 - (base64.length % 4)) % 4),
    "=",
  );
  const binary = window.atob(padded);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i);
  return bytes.buffer;
}

function bufferToBase64Url(value: ArrayBuffer): string {
  const bytes = new Uint8Array(value);
  let binary = "";
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return window
    .btoa(binary)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/g, "");
}

function decodeCredentialCreationOptions(
  options: any,
): PublicKeyCredentialCreationOptions {
  return {
    ...options.publicKey,
    challenge: base64UrlToBuffer(options.publicKey.challenge),
    user: {
      ...options.publicKey.user,
      id: base64UrlToBuffer(options.publicKey.user.id),
    },
    excludeCredentials: options.publicKey.excludeCredentials?.map(
      (credential: any) => ({
        ...credential,
        id: base64UrlToBuffer(credential.id),
      }),
    ),
  };
}

function decodeCredentialRequestOptions(
  options: any,
): PublicKeyCredentialRequestOptions {
  return {
    ...options.publicKey,
    challenge: base64UrlToBuffer(options.publicKey.challenge),
    allowCredentials: options.publicKey.allowCredentials?.map(
      (credential: any) => ({
        ...credential,
        id: base64UrlToBuffer(credential.id),
      }),
    ),
  };
}

export async function createPasskeyCredential(
  options: unknown,
): Promise<string> {
  if (!window.PublicKeyCredential)
    throw new Error("passkeys are not supported by this browser");
  const credential = (await navigator.credentials.create({
    publicKey: decodeCredentialCreationOptions(options),
  })) as PublicKeyCredential | null;
  if (!credential) throw new Error("passkey creation was cancelled");

  const response = credential.response as AuthenticatorAttestationResponse;
  return JSON.stringify({
    id: credential.id,
    rawId: bufferToBase64Url(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64Url(response.clientDataJSON),
      attestationObject: bufferToBase64Url(response.attestationObject),
    },
  });
}

export async function getPasskeyAssertion(options: unknown): Promise<string> {
  if (!window.PublicKeyCredential)
    throw new Error("passkeys are not supported by this browser");
  const credential = (await navigator.credentials.get({
    publicKey: decodeCredentialRequestOptions(options),
  })) as PublicKeyCredential | null;
  if (!credential) throw new Error("passkey login was cancelled");

  const response = credential.response as AuthenticatorAssertionResponse;
  return JSON.stringify({
    id: credential.id,
    rawId: bufferToBase64Url(credential.rawId),
    type: credential.type,
    response: {
      authenticatorData: bufferToBase64Url(response.authenticatorData),
      clientDataJSON: bufferToBase64Url(response.clientDataJSON),
      signature: bufferToBase64Url(response.signature),
      userHandle: response.userHandle
        ? bufferToBase64Url(response.userHandle)
        : null,
    },
  });
}
