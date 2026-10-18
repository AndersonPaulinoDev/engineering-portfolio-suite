import { StyleSheet, Text, View } from 'react-native';

export function StatusBadge({ text }: { text: string }) {
  return (
    <View style={styles.badge}>
      <Text style={styles.text}>{text.replaceAll('_', ' ')}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    backgroundColor: '#202c38',
    paddingHorizontal: 9,
    paddingVertical: 5,
    borderRadius: 999,
  },
  text: {
    color: '#b8c5d0',
    fontSize: 11,
    textTransform: 'capitalize',
  },
});
