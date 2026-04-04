import * as Keychain from 'react-native-keychain';

const SERVICE_NAME = 'com.hellotalk.auth';

export interface Tokens {
  accessToken: string;
  refreshToken: string;
}

export async function saveTokens(tokens: Tokens): Promise<void> {
  await Keychain.setGenericPassword('tokens', JSON.stringify(tokens), {
    service: SERVICE_NAME,
  });
}

export async function getTokens(): Promise<Tokens | null> {
  const credentials = await Keychain.getGenericPassword({service: SERVICE_NAME});
  if (!credentials) {
    return null;
  }
  return JSON.parse(credentials.password) as Tokens;
}

export async function clearTokens(): Promise<void> {
  await Keychain.resetGenericPassword({service: SERVICE_NAME});
}
