import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

const key = 'agentwatch.access';

export async function saveAccessToken(token: string) {
  if (Platform.OS === 'web') {
    localStorage.setItem(key, token);
    return;
  }
  await SecureStore.setItemAsync(key, token);
}

export async function loadAccessToken() {
  if (Platform.OS === 'web') {
    return localStorage.getItem(key);
  }
  return SecureStore.getItemAsync(key);
}

export async function clearAccessToken() {
  if (Platform.OS === 'web') {
    localStorage.removeItem(key);
    return;
  }
  await SecureStore.deleteItemAsync(key);
}
