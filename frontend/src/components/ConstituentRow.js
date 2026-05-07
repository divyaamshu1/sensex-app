import React, { useEffect, useRef } from 'react';
import { formatPoints, formatSigned, formatPct, direction } from '../utils/format';

export default function ConstituentRow({ stock, sensexChange }) {
  const rowRef = useRef(null);
  const prevPoints = useRef(null);

  useEffect(() => {
    if (!rowRef.current) return;
    if (prevPoints.current !== null && prevPoints.current !== stock.pointsChange) {
      const el = rowRef.current;
      const dir = direction(stock.pointsChange);
      el.classList.remove('row-flash-up', 'row-flash-down');
      void el.offsetWidth;
      if (dir !== 'flat') el.classList.add(`row-flash-${dir}`);
    }
    prevPoints.current = stock.pointsChange;
  }, [stock.pointsChange]);

  const dir = direction(stock.pointsChange);
  const isUp = dir === 'up';
  const isDown = dir === 'down';
  const absWeight = stock.weightage;
  // Bar width: proportional to weightage, max at ~12%
  const barPct = Math.min((absWeight / 12) * 100, 100);

  return (
    <div className={`constituent-row dir-${dir}`} ref={rowRef}>
      <div className="row-left">
        <div className="row-rank">{stock.rank}</div>
        <div className="row-info">
          <span className="row-name">{stock.companyName}</span>
          <div className="row-weight-bar-wrap">
            <div className="row-weight-bar" style={{ width: `${barPct}%` }} />
          </div>
        </div>
      </div>

      <div className="row-right">
        <div className="row-metrics">
          <span className="row-weightage">{absWeight.toFixed(2)}%</span>
          <span className={`row-points dir-${dir}`}>
            {isUp ? '▲' : isDown ? '▼' : ''}
            {formatSigned(stock.pointsChange, 2)} pts
          </span>
        </div>
      </div>
    </div>
  );
}
