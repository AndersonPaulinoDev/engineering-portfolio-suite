import { Stack, useLocalSearchParams } from 'expo-router';
import { useEffect, useMemo, useState } from 'react';
import { Pressable, SafeAreaView, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';

import { addNote, getWorkOrder, transition, type Status, type WorkOrder } from '@/lib/api';
import { loadToken } from '@/lib/auth';
import { queueOperation } from '@/lib/offline';

const nextStatus: Partial<Record<Status, Status>> = {
  assigned: 'accepted',
  accepted: 'in_progress',
  blocked: 'in_progress',
  in_progress: 'completed',
};

export default function WorkOrderDetail() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const [token, setToken] = useState<string | null>(null);
  const [workOrder, setWorkOrder] = useState<WorkOrder | null>(null);
  const [note, setNote] = useState('');
  const [message, setMessage] = useState('');

  useEffect(() => {
    loadToken().then(async (saved) => {
      setToken(saved);
      if (saved && id) setWorkOrder(await getWorkOrder(saved, id));
    });
  }, [id]);

  const next = useMemo(() => (workOrder ? nextStatus[workOrder.status] : undefined), [workOrder]);

  async function move() {
    if (!token || !workOrder || !next) return;
    try {
      setWorkOrder(await transition(token, workOrder.id, next));
      setMessage('Status synced.');
    } catch (err) {
      queueOperation({
        id: `op_${Date.now()}`,
        type: 'transition',
        work_order_id: workOrder.id,
        status: next,
      });
      setMessage(`Saved offline: ${(err as Error).message}`);
    }
  }

  async function submitNote() {
    if (!token || !workOrder || !note.trim()) return;
    try {
      setWorkOrder(await addNote(token, workOrder.id, note.trim()));
      setNote('');
      setMessage('Note synced.');
    } catch (err) {
      queueOperation({
        id: `op_${Date.now()}`,
        type: 'note',
        work_order_id: workOrder.id,
        note: note.trim(),
      });
      setNote('');
      setMessage(`Note queued offline: ${(err as Error).message}`);
    }
  }

  if (!workOrder) {
    return <SafeAreaView style={styles.center}><Text>Loading work order…</Text></SafeAreaView>;
  }

  return (
    <SafeAreaView style={styles.root}>
      <Stack.Screen options={{ title: workOrder.title }} />
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.card}>
          <Text style={styles.title}>{workOrder.title}</Text>
          <Text style={styles.meta}>{workOrder.location} · {workOrder.priority}</Text>
          <Text style={styles.status}>{workOrder.status.replaceAll('_', ' ')}</Text>
          <Text style={styles.body}>{workOrder.description}</Text>
        </View>

        <View style={styles.card}>
          <Text style={styles.section}>Instructions</Text>
          <Text style={styles.body}>{workOrder.instructions}</Text>
          <Text style={styles.asset}>Asset: {workOrder.equipment_tag}</Text>
        </View>

        <View style={styles.card}>
          <Text style={styles.section}>Field notes</Text>
          {workOrder.notes.length === 0 && <Text style={styles.muted}>No notes yet.</Text>}
          {workOrder.notes.map((item, index) => <Text key={`${item}-${index}`} style={styles.note}>• {item}</Text>)}
          <TextInput
            value={note}
            onChangeText={setNote}
            placeholder="Add a technician note"
            style={styles.input}
          />
          <Pressable style={styles.secondaryButton} onPress={submitNote}>
            <Text style={styles.secondaryButtonText}>Add note</Text>
          </Pressable>
        </View>

        {!!message && <Text style={styles.message}>{message}</Text>}
        {next && (
          <Pressable style={styles.primaryButton} onPress={move}>
            <Text style={styles.primaryButtonText}>Move to {next.replaceAll('_', ' ')}</Text>
          </Pressable>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#f4f5f7' },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  content: { padding: 16, gap: 12, paddingBottom: 48 },
  card: { backgroundColor: '#fff', padding: 16, borderRadius: 14, gap: 8, borderWidth: 1, borderColor: '#e0e4e8' },
  title: { fontSize: 22, fontWeight: '900', color: '#17202a' },
  section: { fontSize: 16, fontWeight: '800', color: '#17202a' },
  meta: { color: '#6f7b86' },
  status: { alignSelf: 'flex-start', backgroundColor: '#e8f1fd', color: '#315b82', paddingHorizontal: 9, paddingVertical: 5, borderRadius: 999, textTransform: 'capitalize' },
  body: { color: '#34414d', lineHeight: 21 },
  asset: { color: '#47525c', fontFamily: 'monospace' },
  muted: { color: '#7a8792' },
  note: { color: '#34414d' },
  input: { borderWidth: 1, borderColor: '#d7dce1', borderRadius: 10, padding: 12, backgroundColor: '#fafbfc' },
  secondaryButton: { backgroundColor: '#e8f1fd', padding: 12, borderRadius: 10, alignItems: 'center' },
  secondaryButtonText: { color: '#315b82', fontWeight: '800' },
  primaryButton: { backgroundColor: '#1368d4', padding: 15, borderRadius: 12, alignItems: 'center' },
  primaryButtonText: { color: '#fff', fontWeight: '800' },
  message: { color: '#36526d' },
});
