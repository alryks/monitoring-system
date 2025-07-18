import React, { useState, useEffect } from 'react';
import { api } from '../services/api';
import styles from './Domains.module.css';

interface Domain {
  id: string;
  name: string;
  agent_id: string;
  agent_ip: string;
  agent_name?: string;
  is_active: boolean;
  ssl_enabled: boolean;
  created: string;
  updated: string;
  routes: DomainRoute[];
}

interface DomainRoute {
  id: string;
  domain_id: string;
  container_name: string;
  port: string;
  path: string;
  is_active: boolean;
  created: string;
  updated: string;
}

interface Agent {
  id: string;
  name: string;
  public_ip?: string;
}

const Domains: React.FC = () => {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [selectedDomain, setSelectedDomain] = useState<Domain | null>(null);
  const [showRouteForm, setShowRouteForm] = useState(false);

  // Формы
  const [createForm, setCreateForm] = useState({
    name: '',
    agent_id: '',
    ssl_enabled: false
  });

  const [routeForm, setRouteForm] = useState({
    container_name: '',
    port: '',
    path: '/'
  });

  useEffect(() => {
    fetchDomains();
    fetchAgents();
  }, []);

  const fetchDomains = async () => {
    try {
      const response = await api.get('/domains');
      setDomains(response.data.domains || []);
    } catch (error) {
      console.error('Error fetching domains:', error);
      setDomains([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchAgents = async () => {
    try {
      const response = await api.get('/agents');
      setAgents(response.data.agents);
    } catch (error) {
      console.error('Error fetching agents:', error);
    }
  };

  const handleCreateDomain = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.post('/domains', createForm);
      setShowCreateForm(false);
      setCreateForm({ name: '', agent_id: '', ssl_enabled: false });
      fetchDomains();
    } catch (error) {
      console.error('Error creating domain:', error);
    }
  };

  const handleCreateRoute = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedDomain) return;

    try {
      await api.post('/domains/routes', {
        domain_id: selectedDomain.id,
        ...routeForm
      });
      setShowRouteForm(false);
      setRouteForm({ container_name: '', port: '', path: '/' });
      fetchDomains();
    } catch (error) {
      console.error('Error creating route:', error);
    }
  };

  const handleDeleteDomain = async (domainId: string) => {
    if (!confirm('Are you sure you want to delete this domain?')) return;

    try {
      await api.delete(`/domains/${domainId}`);
      fetchDomains();
    } catch (error) {
      console.error('Error deleting domain:', error);
    }
  };

  const handleDeleteRoute = async (routeId: string) => {
    if (!confirm('Are you sure you want to delete this route?')) return;

    try {
      await api.delete(`/domains/routes/${routeId}`);
      fetchDomains();
    } catch (error) {
      console.error('Error deleting route:', error);
    }
  };

  const toggleDomainStatus = async (domain: Domain) => {
    try {
      await api.put(`/domains/${domain.id}`, {
        is_active: !domain.is_active
      });
      fetchDomains();
    } catch (error) {
      console.error('Error updating domain:', error);
    }
  };

  if (loading) {
    return <div className={styles.loading}>Loading domains...</div>;
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1>Domain Management</h1>
        <button 
          className={styles.createButton}
          onClick={() => setShowCreateForm(true)}
        >
          Create Domain
        </button>
      </div>

      {/* Create Domain Form */}
      {showCreateForm && (
        <div className={styles.modal}>
          <div className={styles.modalContent}>
            <h2>Create New Domain</h2>
            <form onSubmit={handleCreateDomain}>
              <div className={styles.formGroup}>
                <label>Domain Name:</label>
                <input
                  type="text"
                  value={createForm.name}
                  onChange={(e) => setCreateForm({...createForm, name: e.target.value})}
                  placeholder="example.com"
                  required
                />
              </div>
              <div className={styles.formGroup}>
                <label>Agent:</label>
                <select
                  value={createForm.agent_id}
                  onChange={(e) => setCreateForm({...createForm, agent_id: e.target.value})}
                  required
                >
                  <option value="">Select Agent</option>
                  {agents.map(agent => (
                    <option key={agent.id} value={agent.id}>
                      {agent.name} ({agent.public_ip})
                    </option>
                  ))}
                </select>
              </div>
              <div className={styles.formGroup}>
                <label>
                  <input
                    type="checkbox"
                    checked={createForm.ssl_enabled}
                    onChange={(e) => setCreateForm({...createForm, ssl_enabled: e.target.checked})}
                  />
                  Enable SSL
                </label>
              </div>
              <div className={styles.formActions}>
                <button type="submit">Create</button>
                <button type="button" onClick={() => setShowCreateForm(false)}>
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Create Route Form */}
      {showRouteForm && selectedDomain && (
        <div className={styles.modal}>
          <div className={styles.modalContent}>
            <h2>Add Route to {selectedDomain.name}</h2>
            <form onSubmit={handleCreateRoute}>
              <div className={styles.formGroup}>
                <label>Container Name:</label>
                <input
                  type="text"
                  value={routeForm.container_name}
                  onChange={(e) => setRouteForm({...routeForm, container_name: e.target.value})}
                  placeholder="my-container"
                  required
                />
              </div>
              <div className={styles.formGroup}>
                <label>Port:</label>
                <input
                  type="text"
                  value={routeForm.port}
                  onChange={(e) => setRouteForm({...routeForm, port: e.target.value})}
                  placeholder="3000"
                  required
                />
              </div>
              <div className={styles.formGroup}>
                <label>Path:</label>
                <input
                  type="text"
                  value={routeForm.path}
                  onChange={(e) => setRouteForm({...routeForm, path: e.target.value})}
                  placeholder="/"
                  required
                />
              </div>
              <div className={styles.formActions}>
                <button type="submit">Add Route</button>
                <button type="button" onClick={() => setShowRouteForm(false)}>
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Domains List */}
      <div className={styles.domainsList}>
        {domains && domains.length > 0 && domains.map(domain => (
          <div key={domain.id} className={styles.domainCard}>
            <div className={styles.domainHeader}>
              <h3>{domain.name}</h3>
              <div className={styles.domainActions}>
                <button
                  className={`${styles.statusButton} ${domain.is_active ? styles.active : styles.inactive}`}
                  onClick={() => toggleDomainStatus(domain)}
                >
                  {domain.is_active ? 'Active' : 'Inactive'}
                </button>
                <button
                  className={styles.addRouteButton}
                  onClick={() => {
                    setSelectedDomain(domain);
                    setShowRouteForm(true);
                  }}
                >
                  Add Route
                </button>
                <button
                  className={styles.deleteButton}
                  onClick={() => handleDeleteDomain(domain.id)}
                >
                  Delete
                </button>
              </div>
            </div>
            
            <div className={styles.domainInfo}>
              <p><strong>Agent:</strong> {domain.agent_name} ({domain.agent_ip})</p>
              <p><strong>SSL:</strong> {domain.ssl_enabled ? 'Enabled' : 'Disabled'}</p>
              <p><strong>Created:</strong> {new Date(domain.created).toLocaleDateString()}</p>
            </div>

            <div className={styles.routesSection}>
              <h4>Routes:</h4>
              {!domain.routes || domain.routes.length === 0 ? (
                <p className={styles.noRoutes}>No routes configured</p>
              ) : (
                <div className={styles.routesList}>
                  {domain.routes.map(route => (
                    <div key={route.id} className={styles.routeItem}>
                      <span className={styles.routePath}>{route.path}</span>
                      <span className={styles.routeTarget}>
                        → {route.container_name}:{route.port}
                      </span>
                      <button
                        className={styles.deleteRouteButton}
                        onClick={() => handleDeleteRoute(route.id)}
                      >
                        ×
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      {domains.length === 0 && (
        <div className={styles.emptyState}>
          <p>No domains configured yet.</p>
          <button onClick={() => setShowCreateForm(true)}>
            Create your first domain
          </button>
        </div>
      )}
    </div>
  );
};

export default Domains; 