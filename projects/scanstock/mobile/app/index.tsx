import { Link } from 'expo-router';
import { useEffect, useMemo, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';

import { listStock, login, syncOps, type Stock } from '@/lib/api';
import { loadToken, saveToken } from '@/lib/auth';
import { all, remove } from '@/lib/offline';

const DEMO_EMAIL = 'stock@scanstock.dev';
const DEMO_PASSWORD = 'demo1234';

export default function Home() {
  const [token, setToken] = useState<string | null>(null);
  const [stock, setStock] = useState<Stock[]>([]);
  const [loading, setLoading] = useState(true);
  const [email, setEmail] = useState(DEMO_EMAIL);
  const [password, setPassword] = useState(DEMO_PASSWORD);
  const [message, setMessage] = useState('');

  useEffect(() => {
    loadToken().then(async (saved) => {
      setToken(saved);
      if (saved) await refresh(saved);
      setLoading(false);
    });
  }, []);

  const reorderCount = useMemo(() => stock.filter((item) => item.needs_reorder).length, [stock]);

  async function refresh(current = token) {
    if (!current) return;
    try {
      setStock(await listStock(current));
      setMessage('');
    } catch (err) {
      setMessage((err as Error).message);
    }
  }

  async function signIn() {
    try {
      const result = await login(email, password);
      await saveToken(result.access_token);
      setToken(result.access_token);
      await refresh(result.access_token);
    } catch (err) {
      setMessage((err as Error).message);
    }
  }

  async function flushQueue() {
    if (!token) return;
    const pending = all();
    if (!pending.length) {
      setMessage('No pending inventory operations.');
      return;
    }
    const results = await syncOps(token, pending);
    for (const result of results) {
      if (result.state === 'synced') remove(result.id);
    }
    setMessage(`${results.filter((item) => item.state === 'synced').length} synced`);
    await refresh(token);
  }

  if (loading) {
    return <SafeAreaView style={styles.center}><ActivityIndicator /></SafeAreaView>;
  }

  if (!token) {
    return (
      <SafeAreaView style={styles.login}>
        <Text style={styles.brand}>ScanStock</Text>
        <Text style={styles.subtitle}>Transactional inventory with an auditable movement ledger.</Text>
        <View style={styles.demoNotice}>
          <Text style={styles.demoTitle}>Demo credentials</Text>
          <Text style={styles.demoText}>Seeded only when the API runs with DEMO_MODE=true.</Text>
        </View>
        <TextInput value={email} onChangeText={setEmail} autoCapitalize="none" style={styles.input} />
        <TextInput value={password} onChangeText={setPassword} secureTextEntry style={styles.input} />
        {!!message && <Text style={styles.error}>{message}</Text>}
        <Pressable style={styles.primary} onPress={signIn}>
          <Text style={styles.primaryText}>Sign in</Text>
        </Pressable>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.root}>
      <View style={styles.header}>
        <View>
          <Text style={styles.brandLight}>ScanStock</Text>
          <Text style={styles.subtitleDark}>{stock.length} SKUs · {reorderCount} need reorder</Text>
        </View>
        <Pressable onPress={flushQueue}><Text style={styles.link}>Sync queue</Text></Pressable>
      </View>
      {!!message && <Text style={styles.message}>{message}</Text>}

      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.actions}>
          <Link href="/scanner" asChild>
            <Pressable style={styles.actionCard}>
              <Text style={styles.actionTitle}>Scan barcode</Text>
              <Text style={styles.actionText}>Find stock by UPC, SKU, QR, or Code 128.</Text>
            </Pressable>
          </Link>
          <Link href="/transactions" asChild>
            <Pressable style={styles.actionCard}>
              <Text style={styles.actionTitle}>Ledger</Text>
              <Text style={styles.actionText}>Inspect inventory movements and idempotency keys.</Text>
            </Pressable>
          </Link>
        </View>

        <Text style={styles.sectionTitle}>Stock on hand</Text>
        {stock.map((item) => (
          <Link key={item.product.id} href={{ pathname: '/products/[id]', params: { id: item.product.id } }} asChild>
            <Pressable style={styles.card}>
              <View style={styles.row}>
                <View style={styles.grow}>
                  <Text style={styles.title}>{item.product.name}</Text>
                  <Text style={styles.meta}>{item.product.sku} · {item.product.location}</Text>
                </View>
                <View>
                  <Text style={styles.quantity}>{item.quantity}</Text>
                  <Text style={item.needs_reorder ? styles.low : styles.healthy}>
                    {item.needs_reorder ? 'reorder' : 'healthy'}
                  </Text>
                </View>
              </View>
            </Pressable>
          </Link>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0d1117' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  login: { flex: 1, justifyContent: 'center', padding: 28, gap: 14, backgroundColor: '#f5f7f9' },
  brand: { fontSize: 30, fontWeight: '900', color: '#0d1117' },
  brandLight: { fontSize: 30, fontWeight: '900', color: '#fff' },
  subtitle: { color: '#596775' },
  subtitleDark: { color: '#8593a3' },
  demoNotice: { backgroundColor: '#fff8dc', padding: 14, borderRadius: 12, gap: 4 },
  demoTitle: { fontWeight: '800', color: '#5f4b00' },
  demoText: { color: '#715d16' },
  input: { backgroundColor: '#fff', borderWidth: 1, borderColor: '#d7dce1', borderRadius: 12, padding: 14 },
  primary: { backgroundColor: '#0d1117', borderRadius: 12, padding: 15, alignItems: 'center' },
  primaryText: { color: '#fff', fontWeight: '800' },
  error: { color: '#b42318' },
  header: { padding: 20, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  link: { color: '#61dafb', fontWeight: '900' },
  message: { color: '#c4d3df', paddingHorizontal: 20, paddingBottom: 8 },
  content: { padding: 16, gap: 12, paddingBottom: 48 },
  actions: { flexDirection: 'row', gap: 10 },
  actionCard: { flex: 1, backgroundColor: '#102d3a', padding: 14, borderRadius: 14, gap: 5 },
  actionTitle: { color: '#61dafb', fontWeight: '900' },
  actionText: { color: '#a9bfca', fontSize: 12, lineHeight: 17 },
  sectionTitle: { color: '#fff', fontSize: 18, fontWeight: '900', marginTop: 4 },
  card: { backgroundColor: '#151b23', borderColor: '#26303b', borderWidth: 1, borderRadius: 15, padding: 16 },
  row: { flexDirection: 'row', justifyContent: 'space-between', gap: 16 },
  grow: { flex: 1 },
  title: { color: '#fff', fontWeight: '800', fontSize: 16 },
  meta: { color: '#8593a3', marginTop: 4 },
  quantity: { color: '#fff', fontSize: 24, fontWeight: '900', textAlign: 'right' },
  low: { color: '#ffad66', fontSize: 11, textAlign: 'right' },
  healthy: { color: '#70d99a', fontSize: 11, textAlign: 'right' },
});
