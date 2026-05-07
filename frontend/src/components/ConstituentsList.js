import React, { useState } from 'react';
import ConstituentRow from './ConstituentRow';

export default function ConstituentsList({ constituents, sensexChange }) {
  const [sortBy, setSortBy] = useState('rank'); // rank | weightage | points

  if (!constituents || constituents.length === 0) {
    return (
      <div className="constituents-empty">
        <div className="spinner" />
        <span>Loading constituents…</span>
      </div>
    );
  }

  const sorted = [...constituents].sort((a, b) => {
    if (sortBy === 'rank') return a.rank - b.rank;
    if (sortBy === 'weightage') return b.weightage - a.weightage;
    if (sortBy === 'points') return Math.abs(b.pointsChange) - Math.abs(a.pointsChange);
    return 0;
  });

  const totalPoints = constituents.reduce((s, c) => s + (c.pointsChange || 0), 0);
  const totalWeight = constituents.reduce((s, c) => s + (c.weightage || 0), 0);

  return (
    <section className="constituents-section">
      <div className="constituents-header">
        <h2 className="section-title">SENSEX 30</h2>
        <div className="sort-tabs">
          {[
            { key: 'rank', label: 'RANK' },
            { key: 'weightage', label: 'WEIGHT' },
            { key: 'points', label: 'IMPACT' },
          ].map(({ key, label }) => (
            <button
              key={key}
              className={`sort-tab ${sortBy === key ? 'active' : ''}`}
              onClick={() => setSortBy(key)}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      <div className="constituents-summary">
        <div className="summary-item">
          <span className="summary-label">Total Weight</span>
          <span className="summary-val">{totalWeight.toFixed(2)}%</span>
        </div>
        <div className="summary-item">
          <span className="summary-label">Attributed Pts</span>
          <span className={`summary-val dir-${totalPoints >= 0 ? 'up' : 'down'}`}>
            {totalPoints >= 0 ? '+' : '−'}{Math.abs(totalPoints).toFixed(2)}
          </span>
        </div>
      </div>

      <div className="constituents-col-labels">
        <span>COMPANY</span>
        <span>WEIGHT · IMPACT</span>
      </div>

      <div className="constituents-list">
        {sorted.map((stock) => (
          <ConstituentRow
            key={stock.scripCode || stock.companyName}
            stock={stock}
            sensexChange={sensexChange}
          />
        ))}
      </div>
    </section>
  );
}
