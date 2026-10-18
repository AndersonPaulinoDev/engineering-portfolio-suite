import { Stack } from 'expo-router';
import { useEffect, useState } from 'react';
import { SafeAreaView, ScrollView, StyleSheet, Text, View } from 'react-native';

import { listTransactions, type InventoryTransaction } from '@/lib/api';
import { loadToken } from '@/lib/auth';

export default function TransactionsScreen() {
  const [transactions, setTransactions] = useState<InventoryTransaction[]>([]);

  useEffect(() => {
    loadToken().then(async (token) => {
      if (token) setTransactions(await listTransactions(token));
    });
  }, []);

  return (
    <SafeAreaView style={styles.root}>
      <Stack.Screen options={{ title: 'Inventory ledger' }} />
      <ScrollView contentContainerStyle={styles.content}>
        {transactions.map((transaction) => (
          <View key={transaction.transaction_id} style={styles.card}>
            <View style={styles.row}>
              <Text style={styles.type}>{transaction.transaction_type.replaceAll('_', ' ')}</Text>
              <Text style={transaction.quantity_change > 0 ? styles.plus : styles.minus}>
                {transaction.quantity_change > 0 ? '+' : ''}{transaction.quantity_change}
              </Text>
            </View>
            <Text style={styles.meta}>{transaction.product_id} · {transaction.device_id}</Text>
            <Text style={styles.key}>{transaction.idempotency_key}</Text>
          </View>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0d1117' },
  content: { padding: 16, gap: 10 },
  card: { backgroundColor: '#151b23', borderWidth: 1, borderColor: '#26303b', padding: 14, borderRadius: 13, gap: 5 },
  row: { flexDirection: 'row', justifyContent: 'space-between' },
  type: { color: '#fff', fontWeight: '800', textTransform: 'capitalize' },
  plus: { color: '#70d99a', fontWeight: '900' },
  minus: { color: '#ff8b8b', fontWeight: '900' },
  meta: { color: '#8593a3' },
  key: { color: '#637486', fontFamily: 'monospace', fontSize: 11 },
});
