import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../services/api';
import styles from './CreateAction.module.css';

interface Agent {
  id: string;
  name: string;
}

interface CreateActionProps {
  onActionCreated: () => void;
}

const CreateAction: React.FC<CreateActionProps> = ({ onActionCreated }) => {
  const { token } = useAuth();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const [agentsLoading, setAgentsLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [mounted, setMounted] = useState(true);

  // Форма
  const [selectedAgent, setSelectedAgent] = useState('');
  const [actionType, setActionType] = useState('');
  const [payload, setPayload] = useState('{}');

  useEffect(() => {
    setMounted(true);
    return () => setMounted(false);
  }, []);

  useEffect(() => {
    if (mounted) {
      fetchAgents();
    }
  }, [mounted, token]);

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
    } finally {
      if (mounted) setAgentsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      let parsedPayload;
      try {
        parsedPayload = JSON.parse(payload);
      } catch (error) {
        alert('Неверный JSON в параметрах');
        setLoading(false);
        return;
      }

      await api.post('/api/actions', {
        agent_id: selectedAgent,
        type: actionType,
        payload: parsedPayload
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });

      // Сброс формы
      setSelectedAgent('');
      setActionType('');
      setPayload('{}');
      setShowForm(false);
      onActionCreated();
    } catch (error) {
      console.error('Error creating action:', error);
      alert('Ошибка при создании действия');
    } finally {
      setLoading(false);
    }
  };

  const getPayloadTemplate = (type: string) => {
    const templates: { [key: string]: string } = {
      'start_container': JSON.stringify({
        image: 'nginx:latest',
        name: 'my-container',
        ports: {
          '80/tcp': '8080'
        },
        environment: {
          'NODE_ENV': 'production'
        },
        domain: 'example.com'
      }, null, 2),
      'stop_container': JSON.stringify({
        container_id: 'container-id-here',
        timeout: 10
      }, null, 2),
      'remove_container': JSON.stringify({
        container_id: 'container-id-here',
        force: false
      }, null, 2),
      'remove_image': JSON.stringify({
        image_id: 'image-id-here',
        force: false
      }, null, 2),
      'write_file': JSON.stringify({
        path: '/path/to/file',
        content: 'File content here',
        mode: 644
      }, null, 2)
    };
    return templates[type] || '{}';
  };

  const handleTypeChange = (type: string) => {
    setActionType(type);
    setPayload(getPayloadTemplate(type));
  };

  if (!showForm) {
    return (
      <div className={styles.container}>
        <button 
          onClick={() => setShowForm(true)}
          className={styles.createButton}
        >
          Создать действие
        </button>
      </div>
    );
  }

  // Проверяем, что agents является массивом
  const agentsArray = Array.isArray(agents) ? agents : [];
  
  // Показываем индикатор загрузки, если агенты еще не загружены
  if (agentsLoading) {
    return (
      <div className={styles.container}>
        <button 
          onClick={() => setShowForm(true)}
          className={styles.createButton}
          disabled
        >
          Загрузка агентов...
        </button>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <h2>Создать новое действие</h2>
      
      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.formGroup}>
          <label htmlFor="agent">Агент:</label>
          <select
            id="agent"
            value={selectedAgent}
            onChange={(e) => setSelectedAgent(e.target.value)}
            required
            className={styles.select}
          >
            <option value="">Выберите агента</option>
            {agentsArray.map(agent => (
              <option key={agent.id} value={agent.id}>
                {agent.name}
              </option>
            ))}
          </select>
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="type">Тип действия:</label>
          <select
            id="type"
            value={actionType}
            onChange={(e) => handleTypeChange(e.target.value)}
            required
            className={styles.select}
          >
            <option value="">Выберите тип действия</option>
            <option value="start_container">Запуск контейнера</option>
            <option value="stop_container">Остановка контейнера</option>
            <option value="remove_container">Удаление контейнера</option>
            <option value="remove_image">Удаление образа</option>
            <option value="restart_nginx">Перезапуск Nginx</option>
            <option value="write_file">Запись в файл</option>
          </select>
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="payload">Параметры (JSON):</label>
          <textarea
            id="payload"
            value={payload}
            onChange={(e) => setPayload(e.target.value)}
            required
            className={styles.textarea}
            rows={10}
            placeholder="Введите JSON параметры..."
          />
        </div>

        <div className={styles.formActions}>
          <button 
            type="submit" 
            disabled={loading}
            className={styles.submitButton}
          >
            {loading ? 'Создание...' : 'Создать действие'}
          </button>
          <button 
            type="button" 
            onClick={() => setShowForm(false)}
            className={styles.cancelButton}
          >
            Отмена
          </button>
        </div>
      </form>
    </div>
  );
};

export default CreateAction; 