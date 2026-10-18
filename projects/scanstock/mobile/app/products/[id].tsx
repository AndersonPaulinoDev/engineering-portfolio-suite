import { Stack, useLocalSearchParams } from 'expo-router';
import { useEffect, useState } from 'react';
import { Pressable, SafeAreaView, ScrollView, StyleSheet, Text, View } from 'react-native';

import { listStock, listTransactions, transact, type InventoryTransaction, type Stock, type TransactionType } from '@/lib/api';
import { loadToken } from '@/lib/auth';
import { queue } from '@/lib/offline';

export default function ProductDetail() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const [token, setToken] = useState<string | null>(null);
  const [stock, setStock] = useState<Stock | null>(null);
  const [transactions, setTransactions] = useState<InventoryTransaction[]>([]);
  const [message, setMessage] = useState('');

  useEffect(() => {
    loadToken().then(async (saved) => {
      setToken(saved);
      if (!saved || !id) return;
      const rows = await listStock(saved);
      setStock(rows.find((row) => row.product.id === id) ?? null);
      setTransactions(await listTransactions(saved, id));
    });
  }, [id]);

  async function adjust(type: TransactionType, quantity = 1) {
    if (!token || !stock) return;
    const operation = {
      id: `mobile_${Date.now()}`,
      product_id: stock.product.id,
      quantity_change: quantity,
      transaction_type: type,
      device_id: 'expo-mobile',
      reference: 'product-detail',
    };

    try {
      await transact(token, { ...operation, idempotency_key: operation.id });
      const rows = await listStock(token);
      setStock(rows.find((row) => row.product.id === stock.product.id) ?? null);
      setTransactions(await listTransactions(token, stock.product.id));
      setMessage('Inventory movement recorded.');
    } catch (err) {
      queue(operation);
      setMessage(`Queued offline: ${(err as Error).message}`);
    }
  }

  if (!stock) {
    return <SafeAreaView style={styles.center}><Text style={styles.light}>Loading product…</Text></SafeAreaView>;
  }

  return (
    <SafeAreaView style={styles.root}>
      <Stack.Screen options={{ title: stock.product.name }} />
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.hero}>
          <Text style={styles.title}>{stock.product.name}</Text>
          <Text style={styles.meta}>{stock.product.sku} · {stock.product.location}</Text>
          <Text style={styles.quantity}>{stock.quantity}</Text>
          <Text style={styles.label}>units on hand</Text>
        </View>

        <View style={styles.actions}>
          <Pressable style={styles.button} onPress={() => adjust('received')}><Text>Receive +1</Text></Pressable>
          <Pressable style={styles.button} onPress={() => adjust('shipped')}><Text>Ship -1</Text></Pressable>
          <Pressable style={styles.button} onPress={() => adjust('damaged')}><Text>Damage -1</Text></Pressable>
        </View>

        {!!message && <Text style={styles.message}>{message}</Text>}

        <Text style={styles.section}>Recent movements</Text>
        {transactions.slice(0, 12).map((transaction) => (
          <View key={transaction.transaction_id} style={styles.card}>
            <View style={styles.row}>
              <Text style={styles.type}>{transaction.transaction_type.replaceAll('_', ' ')}</Text>
              <Text style={transaction.quantity_change > 0 ? styles.plus : styles.minus}>
                {transaction.quantity_change > 0 ? '+' : ''}{transaction.quantity_change}
              </Text>
            </View>
            <Text style={styles.meta}>{transaction.reference || 'No reference'}</Text>
            <Text style={styles.key}>{transaction.idempotency_key}</Text>
          </View>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0d1117' },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: '#0d1117' },
  content: { padding: 16, gap: 12, paddingBottom: 48 },
  hero: { backgroundColor: '#151b23', borderRadius: 16, padding: 18, gap: 5 },
  title: { color: '#fff', fontSize: 22, fontWeight: '900' },
  light: { color: '#fff' },
  meta: { color: '#8593a3' },
  quantity: { color: '#fff', fontSize: 42, fontWeight: '900', marginTop: 10 },
  label: { color: '#8593a3' },
  actions: { flexDirection: 'row', gap: 8, flexWrap: 'wrap' },
  button: { backgroundColor: '#eef3f7', paddingHorizontal: 13, paddingVertical: 10, borderRadius: 10 },
  message: { color: '#c4d3df' },
  section: { color: '#fff', fontSize: 18, fontWeight: '900', marginTop: 6 },
  card: { backgroundColor: '#151b23', borderWidth: 1, borderColor: '#26303b', borderRadius: 13, padding: 14, gap: 5 },
  row: { flexDirection: 'row', justifyContent: 'space-between' },
  type: { color: '#fff', fontWeight: '800', textTransform: 'capitalize' },
  plus: { color: '#70d99a', fontWeight: '900' },
  minus: { color: '#ff8b8b', fontWeight: '900' },
  key: { color: '#637486', fontFamily: 'monospace', fontSize: 11 },
});
