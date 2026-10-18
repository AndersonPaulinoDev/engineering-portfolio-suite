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

import { listWork, login, type WorkOrder } from '@/lib/api';
import { loadToken, saveToken } from '@/lib/auth';
import { pendingOperations, removeOperation } from '@/lib/offline';
import { syncOps } from '@/lib/api';

const DEMO_EMAIL = 'worker@opstrack.dev';
const DEMO_PASSWORD = 'demo1234';

export default function Home() {
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [work, setWork] = useState<WorkOrder[]>([]);
  const [email, setEmail] = useState(DEMO_EMAIL);
  const [password, setPassword] = useState(DEMO_PASSWORD);
  const [error, setError] = useState('');
  const [syncState, setSyncState] = useState('');

  useEffect(() => {
    loadToken().then(async (saved) => {
      setToken(saved);
      if (saved) {
        await refresh(saved);
      }
      setLoading(false);
    });
  }, []);

  const active = useMemo(
    () => work.filter((item) => !['verified', 'cancelled'].includes(item.status)).length,
    [work],
  );

  async function refresh(current = token) {
    if (!current) return;
    try {
      setWork(await listWork(current));
      setError('');
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function signIn() {
    try {
      setError('');
      const result = await login(email, password);
      await saveToken(result.access_token);
      setToken(result.access_token);
      await refresh(result.access_token);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function flushOfflineQueue() {
    if (!token) return;
    const operations = pendingOperations();
    if (!operations.length) {
      setSyncState('No pending offline changes.');
      return;
    }

    const results = await syncOps(token, operations);
    for (const result of results) {
      if (result.state === 'synced') {
        removeOperation(result.id);
      }
    }
    setSyncState(`${results.filter((item) => item.state === 'synced').length} synced`);
    await refresh(token);
  }

  if (loading) {
    return (
      <SafeAreaView style={styles.centered}>
        <ActivityIndicator />
      </SafeAreaView>
    );
  }

  if (!token) {
    return (
      <SafeAreaView style={styles.login}>
        <Text style={styles.brand}>OpsTrack</Text>
        <Text style={styles.subtitle}>Offline-first field operations for technicians and supervisors.</Text>
        <View style={styles.demoNotice}>
          <Text style={styles.demoTitle}>Demo credentials</Text>
          <Text style={styles.demoText}>
            These credentials are seeded only when the API runs with DEMO_MODE=true.
          </Text>
        </View>
        <TextInput
          value={email}
          onChangeText={setEmail}
          autoCapitalize="none"
          keyboardType="email-address"
          style={styles.input}
          placeholder="Email"
        />
        <TextInput
          value={password}
          onChangeText={setPassword}
          secureTextEntry
          style={styles.input}
          placeholder="Password"
        />
        {!!error && <Text style={styles.error}>{error}</Text>}
        <Pressable style={styles.primaryButton} onPress={signIn}>
          <Text style={styles.primaryButtonText}>Sign in</Text>
        </Pressable>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.root}>
      <View style={styles.header}>
        <View>
          <Text style={styles.brand}>OpsTrack</Text>
          <Text style={styles.subtitle}>{active} active work orders</Text>
        </View>
        <Pressable onPress={flushOfflineQueue}>
          <Text style={styles.link}>Sync queue</Text>
        </Pressable>
      </View>

      {!!syncState && <Text style={styles.syncState}>{syncState}</Text>}
      {!!error && <Text style={styles.errorBanner}>{error}</Text>}

      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.quickActions}>
          <Link href="/scanner" asChild>
            <Pressable style={styles.actionCard}>
              <Text style={styles.actionTitle}>Scan equipment</Text>
              <Text style={styles.actionText}>Open a tagged asset and its related work.</Text>
            </Pressable>
          </Link>
          <Link href="/audit" asChild>
            <Pressable style={styles.actionCard}>
              <Text style={styles.actionTitle}>Audit trail</Text>
              <Text style={styles.actionText}>Review immutable operational history.</Text>
            </Pressable>
          </Link>
        </View>

        <Text style={styles.sectionTitle}>Assigned work</Text>
        {work.map((item) => (
          <Link key={item.id} href={{ pathname: '/work-orders/[id]', params: { id: item.id } }} asChild>
            <Pressable style={styles.card}>
              <View style={styles.row}>
                <Text style={styles.cardTitle}>{item.title}</Text>
                <Text style={styles.badge}>{item.status.replaceAll('_', ' ')}</Text>
              </View>
              <Text style={styles.meta}>{item.location} · {item.priority}</Text>
              <Text style={styles.body} numberOfLines={2}>{item.instructions}</Text>
              <Text style={styles.assetTag}>{item.equipment_tag}</Text>
            </Pressable>
          </Link>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#f4f5f7' },
  centered: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  login: { flex: 1, padding: 28, justifyContent: 'center', gap: 14, backgroundColor: '#f4f5f7' },
  brand: { fontSize: 30, fontWeight: '900', color: '#17202a' },
  subtitle: { color: '#6f7b86', marginTop: 2 },
  demoNotice: { backgroundColor: '#fff8dc', padding: 14, borderRadius: 12, gap: 4 },
  demoTitle: { fontWeight: '800', color: '#5f4b00' },
  demoText: { color: '#715d16', lineHeight: 19 },
  input: { backgroundColor: '#fff', borderWidth: 1, borderColor: '#d7dce1', padding: 14, borderRadius: 12 },
  primaryButton: { backgroundColor: '#1368d4', padding: 15, borderRadius: 12, alignItems: 'center' },
  primaryButtonText: { color: '#fff', fontWeight: '800' },
  error: { color: '#b42318' },
  errorBanner: { color: '#b42318', paddingHorizontal: 20, paddingBottom: 8 },
  syncState: { color: '#36526d', paddingHorizontal: 20, paddingBottom: 8 },
  header: { padding: 20, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  link: { color: '#1368d4', fontWeight: '800' },
  content: { padding: 16, gap: 12, paddingBottom: 48 },
  quickActions: { flexDirection: 'row', gap: 10 },
  actionCard: { flex: 1, backgroundColor: '#e8f1fd', padding: 14, borderRadius: 14, gap: 5 },
  actionTitle: { fontWeight: '800', color: '#153a63' },
  actionText: { fontSize: 12, color: '#496b8f', lineHeight: 17 },
  sectionTitle: { fontSize: 18, fontWeight: '900', marginTop: 4, color: '#17202a' },
  card: { backgroundColor: '#fff', padding: 16, borderRadius: 14, gap: 8, borderWidth: 1, borderColor: '#e0e4e8' },
  row: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', gap: 12 },
  cardTitle: { flex: 1, fontSize: 16, fontWeight: '800', color: '#17202a' },
  badge: { fontSize: 11, color: '#315b82', backgroundColor: '#e8f1fd', paddingHorizontal: 8, paddingVertical: 5, borderRadius: 999, textTransform: 'capitalize' },
  meta: { color: '#6f7b86' },
  body: { color: '#34414d', lineHeight: 20 },
  assetTag: { alignSelf: 'flex-start', backgroundColor: '#eef0f2', paddingHorizontal: 8, paddingVertical: 5, borderRadius: 7, color: '#47525c', fontFamily: 'monospace' },
});
