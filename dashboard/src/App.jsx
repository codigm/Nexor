import React, { useEffect, useState } from 'react';
import { io } from 'socket.io-client';

const SOCKET_URL = 'http://localhost:30001';

// ─── Icons ────────────────────────────────────────────
const IconActivity = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
  </svg>
);
const IconSend = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="22" y1="2" x2="11" y2="13" /><polygon points="22 2 15 22 11 13 2 9 22 2" />
  </svg>
);
const IconArrow = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="17 1 21 5 17 9" /><path d="M3 11V9a4 4 0 0 1 4-4h14" /><polyline points="7 23 3 19 7 15" /><path d="M21 13v2a4 4 0 0 1-4 4H3" />
  </svg>
);
const IconLedger = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect x="2" y="3" width="20" height="14" rx="2" /><line x1="8" y1="21" x2="16" y2="21" /><line x1="12" y1="17" x2="12" y2="21" />
  </svg>
);
const IconDollar = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="12" y1="1" x2="12" y2="23" /><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
  </svg>
);
const IconHash = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="4" y1="9" x2="20" y2="9" /><line x1="4" y1="15" x2="20" y2="15" /><line x1="10" y1="3" x2="8" y2="21" /><line x1="16" y1="3" x2="14" y2="21" />
  </svg>
);
const IconZap = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
  </svg>
);
const IconCpu = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect x="4" y="4" width="16" height="16" rx="2" /><rect x="9" y="9" width="6" height="6" /><line x1="9" y1="1" x2="9" y2="4" /><line x1="15" y1="1" x2="15" y2="4" /><line x1="9" y1="20" x2="9" y2="23" /><line x1="15" y1="20" x2="15" y2="23" /><line x1="20" y1="9" x2="23" y2="9" /><line x1="20" y1="14" x2="23" y2="14" /><line x1="1" y1="9" x2="4" y2="9" /><line x1="1" y1="14" x2="4" y2="14" />
  </svg>
);

// Simple SVG Area Chart (no external dep issues)
const MiniChart = ({ data }) => {
  if (!data || data.length < 2) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: 'var(--text-muted)', fontSize: '0.8rem' }}>
      Waiting for data...
    </div>
  );
  const max = Math.max(...data.map(d => d.volume), 1);
  const W = 600; const H = 200;
  const pts = data.map((d, i) => ({
    x: (i / (data.length - 1)) * W,
    y: H - (d.volume / max) * (H - 20)
  }));
  const pathD = pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' ');
  const areaD = `${pathD} L${W},${H} L0,${H} Z`;

  return (
    <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" style={{ width: '100%', height: '100%' }}>
      <defs>
        <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="var(--accent-blue)" stopOpacity="0.25" />
          <stop offset="100%" stopColor="var(--accent-blue)" stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={areaD} fill="url(#areaGrad)" />
      <path d={pathD} fill="none" stroke="var(--accent-blue)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
      {pts.map((p, i) => (
        <circle key={i} cx={p.x} cy={p.y} r="3" fill="var(--accent-blue)" opacity="0.7" />
      ))}
    </svg>
  );
};

const fmt = new Intl.NumberFormat('en-IN', {
  style: 'currency', currency: 'INR', maximumFractionDigits: 0
});

export default function App() {
  const [isConnected, setIsConnected] = useState(false);
  const [transactions, setTransactions] = useState([]);
  const [stats, setStats] = useState({ totalVolume: 0, txCount: 0, rtgsCount: 0, neftCount: 0 });
  const [chartData, setChartData] = useState([]);
  const [formData, setFormData] = useState({
    senderAccount: 'ALPHA001',
    receiverAccount: 'BETA001',
    amount: '15000',
    description: 'Dashboard Transfer'
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [lastTxStatus, setLastTxStatus] = useState(null); // 'success' | 'error'

  useEffect(() => {
    const socket = io(SOCKET_URL);
    socket.on('connect', () => setIsConnected(true));
    socket.on('disconnect', () => setIsConnected(false));
    socket.on('transaction', (tx) => {
      setTransactions(prev => [tx, ...prev].slice(0, 60));
      setStats(prev => ({
        totalVolume: prev.totalVolume + (tx.amount || 0),
        txCount: prev.txCount + 1,
        rtgsCount: prev.rtgsCount + (tx.amount >= 10000 ? 1 : 0),
        neftCount: prev.neftCount + (tx.amount < 10000 ? 1 : 0),
      }));
      setChartData(prev => {
        const now = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        const last = prev[prev.length - 1];
        if (last && last.time === now) {
          return [...prev.slice(0, -1), { ...last, volume: last.volume + (tx.amount || 0) }];
        }
        return [...prev, { time: now, volume: tx.amount || 0 }].slice(-20);
      });
    });
    return () => socket.disconnect();
  }, []);

  const isRTGS = parseFloat(formData.amount) >= 10000;

  const handleTransfer = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setLastTxStatus(null);
    try {
      // Call alpha-bank directly — it publishes to Kafka → entire pipeline processes it
      const res = await fetch('http://localhost:30003/api/transfer', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          senderAccount: formData.senderAccount,
          receiverAccount: formData.receiverAccount,
          amount: parseFloat(formData.amount),
          currency: 'INR',
          description: formData.description
        })
      });
      if (!res.ok) throw new Error('Failed');
      setLastTxStatus('success');
      setTimeout(() => setLastTxStatus(null), 3000);
    } catch {
      setLastTxStatus('error');
      setTimeout(() => setLastTxStatus(null), 3000);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="dashboard-container">

      {/* ── Header ── */}
      <header className="dashboard-header">
        <div className="header-left">
          <h1 className="dashboard-title">
            <IconActivity />
            <span className="title-olympus">OLYMPUS</span>
            <span className="title-sub">Network</span>
          </h1>
          <div className="header-subtitle">Central Bank Real-Time Gross Settlement Platform</div>
        </div>
        <div className="header-right">
          <div className={`status-badge ${isConnected ? '' : 'disconnected'}`}>
            <div className={`status-dot ${isConnected ? 'active' : ''}`}
              style={{ background: isConnected ? 'var(--accent-green)' : 'var(--accent-red)' }} />
            {isConnected ? 'Live Stream' : 'Disconnected'}
          </div>
        </div>
      </header>

      {/* ── Metrics Row ── */}
      <div className="metrics-grid">
        <div className="metric-card green">
          <div className="metric-icon green"><IconDollar /></div>
          <div className="metric-label">Total Settled Volume</div>
          <div className="metric-value green">{fmt.format(stats.totalVolume)}</div>
        </div>
        <div className="metric-card blue">
          <div className="metric-icon blue"><IconHash /></div>
          <div className="metric-label">Transactions Processed</div>
          <div className="metric-value blue">{stats.txCount.toLocaleString()}</div>
        </div>
        <div className="metric-card purple">
          <div className="metric-icon purple"><IconZap /></div>
          <div className="metric-label">RTGS High-Value</div>
          <div className="metric-value purple">{stats.rtgsCount.toLocaleString()}</div>
        </div>
        <div className="metric-card" style={{ '--hover-color': 'rgba(251,191,36,0.15)' }}>
          <div className="metric-icon amber"><IconCpu /></div>
          <div className="metric-label">NEFT Batch</div>
          <div className="metric-value amber">{stats.neftCount.toLocaleString()}</div>
        </div>
      </div>

      {/* ── Content Grid ── */}
      <div className="content-grid">
        <div className="left-column">

          {/* Transfer Form */}
          <div className="glass-card">
            <div className="card-header">
              <div className="card-header-icon"><IconSend /></div>
              Initiate Transfer
              {lastTxStatus === 'success' && (
                <span style={{ marginLeft: 'auto', fontSize: '0.75rem', color: 'var(--accent-green)' }}>✓ Transfer submitted</span>
              )}
              {lastTxStatus === 'error' && (
                <span style={{ marginLeft: 'auto', fontSize: '0.75rem', color: 'var(--accent-red)' }}>✕ Transfer failed — is the API running?</span>
              )}
            </div>
            <form className="transfer-form" onSubmit={handleTransfer}>
              <div className="form-row">
                <div className="form-group">
                  <label>Sender Account</label>
                  <input type="text" className="form-control" value={formData.senderAccount}
                    onChange={e => setFormData({ ...formData, senderAccount: e.target.value })} required />
                </div>
                <div className="form-group">
                  <label>Receiver Account</label>
                  <input type="text" className="form-control" value={formData.receiverAccount}
                    onChange={e => setFormData({ ...formData, receiverAccount: e.target.value })} required />
                </div>
                <div className="form-group" style={{ maxWidth: '140px' }}>
                  <label>Amount (INR)</label>
                  <input type="number" className="form-control" value={formData.amount}
                    onChange={e => setFormData({ ...formData, amount: e.target.value })} required min="1" />
                </div>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>Description / Remarks</label>
                  <input type="text" className="form-control" value={formData.description}
                    onChange={e => setFormData({ ...formData, description: e.target.value })} />
                </div>
              </div>
              <div className="transfer-footer">
                <span className={`tx-type-badge ${isRTGS ? 'rtgs' : 'neft'}`}>
                  {isRTGS ? '⚡ RTGS' : '📦 NEFT'} — {isRTGS ? 'Immediate high-value settlement' : 'Batched settlement (< ₹10,000)'}
                </span>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? 'Processing…' : 'Execute Transfer →'}
                </button>
              </div>
            </form>
          </div>

          {/* Chart */}
          <div className="glass-card" style={{ flex: 1 }}>
            <div className="card-header">
              <div className="card-header-icon"><IconActivity /></div>
              Network Throughput — Volume (INR)
            </div>
            <div className="chart-wrapper">
              <MiniChart data={chartData} />
            </div>
          </div>
        </div>

        {/* Live Ledger */}
        <div className="glass-card">
          <div className="card-header">
            <div className="card-header-icon"><IconLedger /></div>
            Live Settlement Ledger
            <span style={{ marginLeft: 'auto', fontSize: '0.7rem', color: 'var(--text-muted)' }}>
              {transactions.length > 0 ? `${transactions.length} entries` : ''}
            </span>
          </div>
          <div className="feed-container">
            {transactions.length === 0 ? (
              <div className="feed-empty">
                <IconLedger />
                <span>Awaiting transactions…</span>
              </div>
            ) : (
              transactions.map((tx, idx) => {
                const rtgs = tx.amount >= 10000;
                return (
                  <div key={idx} className="tx-row">
                    <div className="tx-left">
                      <div className={`tx-icon-wrapper ${rtgs ? 'rtgs' : 'neft'}`}>
                        <IconArrow />
                      </div>
                      <div className="tx-details">
                        <div className="tx-route">
                          {tx.senderAccount || '?'} → {tx.receiverAccount || '?'}
                        </div>
                        <div className="tx-desc">{tx.description || 'Inter-bank Transfer'}</div>
                      </div>
                    </div>
                    <div className="tx-right">
                      <div className="tx-amount">{fmt.format(tx.amount || 0)}</div>
                      <span className={`tx-badge ${rtgs ? 'rtgs-badge' : 'neft-badge'}`}>
                        {rtgs ? 'RTGS' : 'NEFT'}
                      </span>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
