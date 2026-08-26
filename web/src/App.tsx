import { useState, useEffect } from 'react'
import './App.css'

interface Model {
  id: string
  object: string
  owned_by: string
}

function App() {
  const [models, setModels] = useState<Model[]>([])
  const [health, setHealth] = useState<string>('checking...')

  useEffect(() => {
    // 检查健康状态
    fetch('/health')
      .then(res => res.json())
      .then(data => setHealth(data.status))
      .catch(() => setHealth('error'))

    // 获取模型列表
    fetch('/v1/models')
      .then(res => res.json())
      .then(data => setModels(data.data || []))
      .catch(() => setModels([]))
  }, [])

  return (
    <div className="app">
      {/* 导航栏 */}
      <nav className="navbar">
        <div className="navbar-brand">
          <span className="logo">⚡</span>
          <span className="brand-name">Token Hub</span>
        </div>
        <div className="navbar-menu">
          <a href="#features">功能</a>
          <a href="#models">模型</a>
          <a href="#docs">文档</a>
        </div>
      </nav>

      {/* 主内容 */}
      <main className="main-content">
        {/* Hero 区域 */}
        <section className="hero">
          <h1>Token Hub</h1>
          <p className="hero-subtitle">下一代 LLM 网关和 AI 资产管理系统</p>
          <p className="hero-description">
            统一管理多种 AI 模型 API，提供用户管理、计费、速率限制等功能
          </p>
          <div className="hero-status">
            <span className={`status-dot ${health === 'ok' ? 'online' : 'offline'}`}></span>
            <span>服务状态: {health}</span>
          </div>
        </section>

        {/* 功能特性 */}
        <section id="features" className="features">
          <h2>核心功能</h2>
          <div className="features-grid">
            <div className="feature-card">
              <div className="feature-icon">🔌</div>
              <h3>统一 API 网关</h3>
              <p>聚合多种 AI 模型 API，提供统一的访问接口</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">👥</div>
              <h3>用户管理</h3>
              <p>多用户支持，权限控制，Token 管理</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">💰</div>
              <h3>计费系统</h3>
              <p>灵活的计费策略，使用量统计</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">⚡</div>
              <h3>速率限制</h3>
              <p>用户级模型限流，防止滥用</p>
            </div>
          </div>
        </section>

        {/* 模型列表 */}
        <section id="models" className="models">
          <h2>可用模型</h2>
          <div className="models-list">
            {models.length > 0 ? (
              models.map(model => (
                <div key={model.id} className="model-card">
                  <div className="model-icon">🤖</div>
                  <div className="model-info">
                    <h3>{model.id}</h3>
                    <p>提供商: {model.owned_by}</p>
                  </div>
                </div>
              ))
            ) : (
              <p className="no-models">暂无可用模型</p>
            )}
          </div>
        </section>

        {/* 快速开始 */}
        <section className="quick-start">
          <h2>快速开始</h2>
          <div className="code-block">
            <pre>
              <code>
{`# 获取模型列表
curl http://localhost:3000/v1/models

# 健康检查
curl http://localhost:3000/health`}
              </code>
            </pre>
          </div>
        </section>
      </main>

      {/* 底部 */}
      <footer className="footer">
        <p>© 2026 Token Hub. 基于 Go + Gin + React 构建</p>
      </footer>
    </div>
  )
}

export default App
