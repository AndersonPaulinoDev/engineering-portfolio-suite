import { StyleSheet, Text, View } from 'react-native';

export function Metric({ label, value }: { label: string; value: number }) {
  return (
    <View style={styles.metric}>
      <Text style={styles.value}>{value}</Text>
      <Text style={styles.label}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  metric: {
    flex: 1,
    backgroundColor: '#111821',
    padding: 16,
    borderRadius: 14,
  },
  value: {
    color: '#fff',
    fontSize: 26,
    fontWeight: '800',
  },
  label: {
    color: '#8698a8',
  },
});
