import { Stack } from 'expo-router';
import { useEffect, useState } from 'react';
import { SafeAreaView, ScrollView, StyleSheet, Text, View } from 'react-native';

import { listAudit, type AuditEntry } from '@/lib/api';
import { loadToken } from '@/lib/auth';

export default function AuditScreen() {
  const [entries, setEntries] = useState<AuditEntry[]>([]);

  useEffect(() => {
    loadToken().then(async (token) => {
      if (token) setEntries(await listAudit(token));
    });
  }, []);

  return (
    <SafeAreaView style={styles.root}>
      <Stack.Screen options={{ title: 'Audit trail' }} />
      <ScrollView contentContainerStyle={styles.content}>
        {entries.map((entry) => (
          <View key={entry.id} style={styles.card}>
            <Text style={styles.action}>{entry.action}</Text>
            <Text style={styles.meta}>{entry.work_order} · {entry.actor}</Text>
            <Text style={styles.time}>{new Date(entry.timestamp).toLocaleString()}</Text>
            {entry.metadata && <Text style={styles.metadata}>{JSON.stringify(entry.metadata)}</Text>}
          </View>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#f4f5f7' },
  content: { padding: 16, gap: 10 },
  card: { backgroundColor: '#fff', padding: 14, borderRadius: 12, borderWidth: 1, borderColor: '#e0e4e8', gap: 4 },
  action: { fontWeight: '800', color: '#17202a' },
  meta: { color: '#52606c' },
  time: { color: '#7a8792', fontSize: 12 },
  metadata: { fontFamily: 'monospace', fontSize: 12, color: '#4b5965' },
});
