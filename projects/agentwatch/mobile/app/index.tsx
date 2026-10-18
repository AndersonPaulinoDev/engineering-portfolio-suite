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

import { Metric } from '@/components/Metric';
import { StatusBadge } from '@/components/StatusBadge';
import {
  decideApproval,
  listAgents,
  listApprovals,
  listProjects,
  listTasks,
  login,
  websocketUrl,
  type Agent,
  type Approval,
  type Project,
  type Task,
} from '@/lib/api';
import { loadAccessToken, saveAccessToken } from '@/lib/auth';

type Tab = 'overview' | 'agents' | 'tasks' | 'approvals';

const DEMO_EMAIL = 'demo@agentwatch.dev';
const DEMO_PASSWORD = 'demo1234';

export default function Home() {
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [email, setEmail] = useState(DEMO_EMAIL);
  const [password, setPassword] = useState(DEMO_PASSWORD);
  const [error, setError] = useState('');

  const [projects, setProjects] = useState<Project[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [approvals, setApprovals] = useState<Approval[]>([]);
  const [tab, setTab] = useState<Tab>('overview');
  const [realtimeEvents, setRealtimeEvents] = useState(0);

  useEffect(() => {
    loadAccessToken().then(async (savedToken) => {
      setToken(savedToken);
      if (savedToken) {
        await refresh(savedToken);
      }
      setLoading(false);
    });
  }, []);

  useEffect(() => {
    if (!token) return;

    const socket = new WebSocket(websocketUrl(token));
    socket.onmessage = () => {
      setRealtimeEvents((count) => count + 1);
      void refresh(token);
    };
    socket.onerror = () => setError('Realtime connection interrupted. HTTP data is still available.');

    return () => socket.close();
  }, [token]);

  const activeAgents = useMemo(
    () => agents.filter((agent) => ['running', 'waiting', 'needs_approval'].includes(agent.state)).length,
    [agents],
  );

  const openTasks = useMemo(
    () => tasks.filter((task) => !['completed', 'cancelled'].includes(task.status)).length,
    [tasks],
  );

  async function refresh(currentToken = token) {
    if (!currentToken) return;

    try {
      const [projectRows, agentRows, taskRows, approvalRows] = await Promise.all([
        listProjects(currentToken),
        listAgents(currentToken),
        listTasks(currentToken),
        listApprovals(currentToken),
      ]);
      setProjects(projectRows);
      setAgents(agentRows);
      setTasks(taskRows);
      setApprovals(approvalRows);
      setError('');
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function signIn() {
    try {
      setError('');
      const result = await login(email, password);
      await saveAccessToken(result.tokens.access_token);
      setToken(result.tokens.access_token);
      await refresh(result.tokens.access_token);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function resolveApproval(id: string, decision: 'approved' | 'rejected') {
    if (!token) return;
    try {
      await decideApproval(token, id, decision);
      await refresh(token);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  if (loading) {
    return (
      <SafeAreaView style={styles.center}>
        <ActivityIndicator />
      </SafeAreaView>
    );
  }

  if (!token) {
    return (
      <SafeAreaView style={styles.login}>
        <Text style={styles.brand}>AgentWatch</Text>
        <Text style={styles.subtitle}>Realtime control center for coding agents and background work.</Text>
        <View style={styles.demoNotice}>
          <Text style={styles.demoTitle}>Demo credentials</Text>
          <Text style={styles.demoText}>
            The account below exists only when the API is started with DEMO_MODE=true.
          </Text>
        </View>
        <TextInput
          value={email}
          onChangeText={setEmail}
          autoCapitalize="none"
          keyboardType="email-address"
          style={styles.input}
          placeholder="Email"
          placeholderTextColor="#66788a"
        />
        <TextInput
          value={password}
          onChangeText={setPassword}
          secureTextEntry
          style={styles.input}
          placeholder="Password"
          placeholderTextColor="#66788a"
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
          <Text style={styles.brand}>AgentWatch</Text>
          <Text style={styles.subtitle}>
            {activeAgents} active agents · {realtimeEvents} realtime updates
          </Text>
        </View>
        <Pressable onPress={() => refresh()}>
          <Text style={styles.link}>Refresh</Text>
        </Pressable>
      </View>

      {!!error && <Text style={styles.errorBanner}>{error}</Text>}

      <View style={styles.tabs}>
        {(['overview', 'agents', 'tasks', 'approvals'] as Tab[]).map((item) => (
          <Pressable
            key={item}
            onPress={() => setTab(item)}
            style={[styles.tab, tab === item && styles.tabActive]}
          >
            <Text style={tab === item ? styles.tabTextActive : styles.tabText}>{item}</Text>
          </Pressable>
        ))}
      </View>

      <ScrollView contentContainerStyle={styles.content}>
        {tab === 'overview' && (
          <>
            <View style={styles.metrics}>
              <Metric label="Projects" value={projects.length} />
              <Metric label="Agents" value={agents.length} />
              <Metric label="Open tasks" value={openTasks} />
            </View>
            <Text style={styles.section}>Recent projects</Text>
            {projects.map((project) => (
              <View key={project.id} style={styles.card}>
                <Text style={styles.title}>{project.name}</Text>
                <Text style={styles.muted}>{project.id}</Text>
              </View>
            ))}
          </>
        )}

        {tab === 'agents' && agents.map((agent) => (
          <View key={agent.id} style={styles.card}>
            <View style={styles.row}>
              <Text style={styles.title}>{agent.name}</Text>
              <StatusBadge text={agent.state} />
            </View>
            <Text style={styles.muted}>{agent.type} · {agent.latest_message}</Text>
          </View>
        ))}

        {tab === 'tasks' && tasks.map((task) => (
          <View key={task.id} style={styles.card}>
            <View style={styles.row}>
              <Text style={styles.title}>{task.title}</Text>
              <StatusBadge text={task.status} />
            </View>
            <Text style={styles.muted}>{task.description}</Text>
          </View>
        ))}

        {tab === 'approvals' && approvals.map((approval) => (
          <View key={approval.id} style={styles.card}>
            <View style={styles.row}>
              <Text style={styles.title}>{approval.action}</Text>
              <StatusBadge text={approval.status} />
            </View>
            <Text style={styles.muted}>{approval.reason}</Text>
            {approval.status === 'pending' && (
              <View style={styles.approvalActions}>
                <Pressable
                  style={styles.rejectButton}
                  onPress={() => resolveApproval(approval.id, 'rejected')}
                >
                  <Text style={styles.rejectText}>Reject</Text>
                </Pressable>
                <Pressable
                  style={styles.approveButton}
                  onPress={() => resolveApproval(approval.id, 'approved')}
                >
                  <Text style={styles.primaryButtonText}>Approve</Text>
                </Pressable>
              </View>
            )}
          </View>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0b0f14' },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  login: { flex: 1, backgroundColor: '#0b0f14', padding: 28, justifyContent: 'center', gap: 12 },
  brand: { color: '#f3f7fb', fontSize: 28, fontWeight: '800' },
  subtitle: { color: '#8da0b4', fontSize: 14, lineHeight: 20 },
  demoNotice: { backgroundColor: '#1a2114', borderColor: '#394821', borderWidth: 1, borderRadius: 12, padding: 13, gap: 4 },
  demoTitle: { color: '#d7ff47', fontWeight: '800' },
  demoText: { color: '#aebb91', lineHeight: 18 },
  input: { backgroundColor: '#151c24', color: '#fff', padding: 14, borderRadius: 12, borderWidth: 1, borderColor: '#25313d' },
  primaryButton: { backgroundColor: '#d7ff47', padding: 15, borderRadius: 12, alignItems: 'center' },
  primaryButtonText: { color: '#0b0f14', fontWeight: '800' },
  error: { color: '#ff8080' },
  errorBanner: { color: '#ff9b9b', paddingHorizontal: 20, paddingBottom: 8 },
  header: { padding: 20, paddingTop: 24, flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  link: { color: '#d7ff47', fontWeight: '700' },
  tabs: { flexDirection: 'row', paddingHorizontal: 12, gap: 4 },
  tab: { paddingVertical: 10, paddingHorizontal: 12, borderRadius: 9 },
  tabActive: { backgroundColor: '#1d2833' },
  tabText: { color: '#718294', textTransform: 'capitalize' },
  tabTextActive: { color: '#fff', fontWeight: '700', textTransform: 'capitalize' },
  content: { padding: 16, gap: 10, paddingBottom: 48 },
  metrics: { flexDirection: 'row', gap: 10 },
  section: { color: '#fff', fontSize: 18, fontWeight: '800', marginTop: 12 },
  card: { backgroundColor: '#111821', padding: 16, borderRadius: 14, borderWidth: 1, borderColor: '#1c2833', gap: 6 },
  row: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 12 },
  title: { color: '#f7fafc', fontWeight: '800', fontSize: 16, flexShrink: 1 },
  muted: { color: '#8698a8' },
  approvalActions: { flexDirection: 'row', justifyContent: 'flex-end', gap: 8, marginTop: 10 },
  rejectButton: { backgroundColor: '#26313b', paddingHorizontal: 16, paddingVertical: 10, borderRadius: 10 },
  rejectText: { color: '#f1f5f8', fontWeight: '700' },
  approveButton: { backgroundColor: '#d7ff47', paddingHorizontal: 16, paddingVertical: 10, borderRadius: 10 },
});
