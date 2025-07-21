import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Agents from './pages/Agents'
import Containers from './pages/Containers'
import Images from './pages/Images'
import Networks from './pages/Networks'
import Notifications from './pages/Notifications'
import Domains from './pages/Domains'
import AgentDetail from './pages/AgentDetail'
import ContainerDetail from './pages/ContainerDetail'
import Layout from './components/Layout'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token } = useAuth()
  
  if (!token) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

function AppRoutes() {
  const { token } = useAuth()

  if (!token) {
    return (
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route 
          path="/dashboard" 
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/agents" 
          element={
            <ProtectedRoute>
              <Agents />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/agents/:id" 
          element={
            <ProtectedRoute>
              <AgentDetail />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/containers" 
          element={
            <ProtectedRoute>
              <Containers />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/containers/:id" 
          element={
            <ProtectedRoute>
              <ContainerDetail />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/images" 
          element={
            <ProtectedRoute>
              <Images />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/domains" 
          element={
            <ProtectedRoute>
              <Domains />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/networks" 
          element={
            <ProtectedRoute>
              <Networks />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/notifications" 
          element={
            <ProtectedRoute>
              <Notifications />
            </ProtectedRoute>
          } 
        />
        <Route path="/login" element={<Navigate to="/dashboard" replace />} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </Layout>
  )
}

function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  )
}

export default App
