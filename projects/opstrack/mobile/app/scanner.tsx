import { Stack, router } from 'expo-router';
import { CameraView, useCameraPermissions } from 'expo-camera';
import { useState } from 'react';
import { Pressable, SafeAreaView, StyleSheet, Text, View } from 'react-native';

import { byEquipment } from '@/lib/api';
import { loadToken } from '@/lib/auth';

export default function ScannerScreen() {
  const [permission, requestPermission] = useCameraPermissions();
  const [message, setMessage] = useState('Point the camera at an equipment QR or barcode.');
  const [locked, setLocked] = useState(false);

  async function handleScan(data: string) {
    if (locked) return;
    setLocked(true);
    try {
      const token = await loadToken();
      if (!token) throw new Error('Sign in first.');
      const matches = await byEquipment(token, data);
      if (!matches.length) {
        setMessage(`No work orders found for ${data}`);
        setLocked(false);
        return;
      }
      router.replace({ pathname: '/work-orders/[id]', params: { id: matches[0].id } });
    } catch (err) {
      setMessage((err as Error).message);
      setLocked(false);
    }
  }

  if (!permission?.granted) {
    return (
      <SafeAreaView style={styles.center}>
        <Stack.Screen options={{ title: 'Equipment scanner' }} />
        <Text style={styles.message}>Camera permission is required to scan equipment tags.</Text>
        <Pressable style={styles.button} onPress={requestPermission}>
          <Text style={styles.buttonText}>Grant camera access</Text>
        </Pressable>
      </SafeAreaView>
    );
  }

  return (
    <View style={styles.root}>
      <Stack.Screen options={{ title: 'Equipment scanner' }} />
      <CameraView
        style={StyleSheet.absoluteFill}
        barcodeScannerSettings={{ barcodeTypes: ['qr', 'ean13', 'code128'] }}
        onBarcodeScanned={({ data }) => handleScan(data)}
      />
      <View style={styles.overlay}>
        <Text style={styles.overlayText}>{message}</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#000' },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 14, padding: 28 },
  message: { textAlign: 'center', color: '#34414d' },
  button: { backgroundColor: '#1368d4', paddingHorizontal: 18, paddingVertical: 13, borderRadius: 10 },
  buttonText: { color: '#fff', fontWeight: '800' },
  overlay: { position: 'absolute', left: 20, right: 20, bottom: 40, backgroundColor: 'rgba(0,0,0,0.72)', padding: 14, borderRadius: 12 },
  overlayText: { color: '#fff', textAlign: 'center' },
});
