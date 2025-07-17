import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../services/api';
import CreateAction from '../components/CreateAction';
import styles from './Actions.module.css';

interface Action {
  id: string;
  agent_id: string;
  type: string;
  payload: any;
  status: string;
  created: string;
  completed?: string;
  response?: string;
  error?: string;
}

interface Agent {
  id: string;
  name: string;
}

const Actions: React.FC = () => {
  const { token } = useAuth();
  const [actions, setActions] = useState<Action[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedAgent, setSelectedAgent] = useState<string>('');
  const [selectedType, setSelectedType] = useState<string>('');
  const [selectedStatus, setSelectedStatus] = useState<string>('');
  const [mounted, setMounted] = useState(true);

  useEffect(() => {
    setMounted(true);
    return () => setMounted(false);
  }, []);

  useEffect(() => {
    if (mounted) {
      fetchActions();
      fetchAgents();
    }
  }, [selectedAgent, selectedType, selectedStatus, token, mounted]);

  const fetchActions = async () => {
    try {
      const params = new URLSearchParams();
      if (selectedAgent) params.append('agent_id', selectedAgent);
      if (selectedType) params.append('type', selectedType);
      if (selectedStatus) params.append('status', selectedStatus);

      const response = await api.get(`/api/actions?${params.toString()}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      

      
      // Проверяем, что response.data.actions существует и является массивом
      if (response.data && response.data.actions && Array.isArray(response.data.actions)) {
        if (mounted) setActions(response.data.actions);
      } else if (response.data && Array.isArray(response.data)) {
        // Если API возвращает массив напрямую
        if (mounted) setActions(response.data);
      } else if (response.data && response.data.actions === null) {
        // Если API возвращает null для actions
        if (mounted) setActions([]);
      } else {
        console.warn('Invalid actions data format:', response.data);
        if (mounted) setActions([]);
      }
    } catch (error) {
      console.error('Error fetching actions:', error);
      if (mounted) setActions([]);
    } finally {
      if (mounted) setLoading(false);
    }
  };

  const fetchAgents = async () => {
    try {
      const response = await api.get('/api/agents', {
        headers: { Authorization: `Bearer ${token}` }
      });
      

      
      // Проверяем, что response.data существует и является массивом
      if (response.data && Array.isArray(response.data)) {
        if (mounted) setAgents(response.data);
      } else {
        console.warn('Invalid agents data format:', response.data);
        if (mounted) setAgents([]);
      }
    } catch (error) {
      console.error('Error fetching agents:', error);
      if (mounted) setAgents([]);
    }
  };

  const getActionTypeLabel = (type: string) => {
    const labels: { [key: string]: string } = {
      'start_container': 'Запуск контейнера',
      'stop_container': 'Остановка контейнера',
      'remove_container': 'Удаление контейнера',
      'remove_image': 'Удаление образа',
      'restart_nginx': 'Перезапуск Nginx',
      'write_file': 'Запись в файл'
    };
    return labels[type] || type;
  };

  const getStatusLabel = (status: string) => {
    const labels: { [key: string]: string } = {
      'pending': 'Ожидает',
      'completed': 'Завершено',
      'failed': 'Ошибка'
    };
    return labels[status] || status;
  };

  const getStatusClass = (status: string) => {
    const classes: { [key: string]: string } = {
      'pending': styles.statusPending,
      'completed': styles.statusCompleted,
      'failed': styles.statusFailed
    };
    return classes[status] || '';
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('ru-RU');
  };

  if (loading) {
    return <div className={styles.loading}>Загрузка...</div>;
  }

  return (
    <div className={styles.container}>
      <h1>Действия</h1>
      
      <CreateAction onActionCreated={fetchActions} />
      
      <div className={styles.filters}>
        <select
          value={selectedAgent}
          onChange={(e) => setSelectedAgent(e.target.value)}
          className={styles.filter}
        >
          <option value="">Все агенты</option>
          {Array.isArray(agents) && agents.map(agent => (
            <option key={agent.id} value={agent.id}>
              {agent.name}
            </option>
          ))}
        </select>

        <select
          value={selectedType}
          onChange={(e) => setSelectedType(e.target.value)}
          className={styles.filter}
        >
          <option value="">Все типы</option>
          <option value="start_container">Запуск контейнера</option>
          <option value="stop_container">Остановка контейнера</option>
          <option value="remove_container">Удаление контейнера</option>
          <option value="remove_image">Удаление образа</option>
          <option value="restart_nginx">Перезапуск Nginx</option>
          <option value="write_file">Запись в файл</option>
        </select>

        <select
          value={selectedStatus}
          onChange={(e) => setSelectedStatus(e.target.value)}
          className={styles.filter}
        >
          <option value="">Все статусы</option>
          <option value="pending">Ожидает</option>
          <option value="completed">Завершено</option>
          <option value="failed">Ошибка</option>
        </select>
      </div>

      <div className={styles.actionsList}>
        {!Array.isArray(actions) || actions.length === 0 ? (
          <div className={styles.noActions}>Действия не найдены</div>
        ) : (
          actions.map(action => (
            <div key={action.id} className={styles.actionCard}>
              <div className={styles.actionHeader}>
                <h3>{getActionTypeLabel(action.type)}</h3>
                <span className={`${styles.status} ${getStatusClass(action.status)}`}>
                  {getStatusLabel(action.status)}
                </span>
              </div>
              
              <div className={styles.actionDetails}>
                <p><strong>ID:</strong> {action.id}</p>
                <p><strong>Агент:</strong> {Array.isArray(agents) && agents.find(a => a.id === action.agent_id)?.name || action.agent_id}</p>
                <p><strong>Создано:</strong> {formatDate(action.created)}</p>
                {action.completed && (
                  <p><strong>Завершено:</strong> {formatDate(action.completed)}</p>
                )}
                
                {action.response && (
                  <div className={styles.response}>
                    <strong>Ответ:</strong>
                    <pre>{action.response}</pre>
                  </div>
                )}
                
                {action.error && (
                  <div className={styles.error}>
                    <strong>Ошибка:</strong>
                    <pre>{action.error}</pre>
                  </div>
                )}
                
                <div className={styles.payload}>
                  <strong>Параметры:</strong>
                  <pre>{JSON.stringify(action.payload, null, 2)}</pre>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Actions; 