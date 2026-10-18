import { Stack, router } from 'expo-router';
import { CameraView, useCameraPermissions } from 'expo-camera';
import { useState } from 'react';
import { Pressable, SafeAreaView, StyleSheet, Text, View } from 'react-native';

import { byCode } from '@/lib/api';
import { loadToken } from '@/lib/auth';

export default function ScannerScreen() {
  const [permission, requestPermission] = useCameraPermissions();
  const [message, setMessage] = useState('Scan a product barcode.');
  const [locked, setLocked] = useState(false);

  async function handleScan(code: string) {
    if (locked) return;
    setLocked(true);
    try {
      const token = await loadToken();
      if (!token) throw new Error('Sign in first.');
      const stock = await byCode(token, code);
      router.replace({ pathname: '/products/[id]', params: { id: stock.product.id } });
    } catch (err) {
      setMessage((err as Error).message);
      setLocked(false);
    }
  }

  if (!permission?.granted) {
    return (
      <SafeAreaView style={styles.center}>
        <Stack.Screen options={{ title: 'Barcode scanner' }} />
        <Text>{message}</Text>
        <Pressable style={styles.button} onPress={requestPermission}>
          <Text style={styles.buttonText}>Grant camera access</Text>
        </Pressable>
      </SafeAreaView>
    );
  }

  return (
    <View style={styles.root}>
      <Stack.Screen options={{ title: 'Barcode scanner' }} />
      <CameraView
        style={StyleSheet.absoluteFill}
        barcodeScannerSettings={{ barcodeTypes: ['ean13', 'upc_a', 'code128', 'qr'] }}
        onBarcodeScanned={({ data }) => handleScan(data)}
      />
      <View style={styles.overlay}><Text style={styles.overlayText}>{message}</Text></View>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#000' },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 14, padding: 28 },
  button: { backgroundColor: '#0d1117', paddingHorizontal: 18, paddingVertical: 13, borderRadius: 10 },
  buttonText: { color: '#fff', fontWeight: '800' },
  overlay: { position: 'absolute', left: 20, right: 20, bottom: 40, backgroundColor: 'rgba(0,0,0,0.72)', padding: 14, borderRadius: 12 },
  overlayText: { color: '#fff', textAlign: 'center' },
});
