import React, { useEffect, useRef } from 'react';
import { formatPoints, formatSigned, formatPct, direction } from '../utils/format';

export default function SensexHeader({ index, status, lastUpdated }) {
  const prevRef = useRef(null);
  const valueRef = useRef(null);

  // Flash animation when value changes
  useEffect(() => {
    if (!index || !valueRef.current) return;
    const el = valueRef.current;
    el.classList.remove('flash-up', 'flash-down');
    void el.offsetWidth; // reflow
    const dir = direction(index.change);
    if (dir !== 'flat') el.classList.add(`flash-${dir}`);
    prevRef.current = index.last;
  }, [index?.last]);

  const dir = index ? direction(index.change) : 'flat';
  const isUp = dir === 'up';
  const isDown = dir === 'down';

  return (
    <header className="sensex-header">
      <div className="header-top">
        <div className="header-label">
          <span className="exchange-tag">BSE</span>
          <span className="index-name">SENSEX</span>
        </div>
        <div className={`status-pill status-${status}`}>
          <span className="status-dot" />
          {status === 'live' ? 'LIVE' : status === 'connecting' ? 'CONNECTING' : 'ERROR'}
        </div>
      </div>

      <div className="header-main" ref={valueRef}>
        <span className="sensex-value">
          {index ? formatPoints(index.last) : '——,———.——'}
        </span>
      </div>

      <div className="header-change-row">
        <span className={`change-badge change-${dir}`}>
          <span className="change-arrow">{isUp ? '▲' : isDown ? '▼' : '—'}</span>
          <span className="change-pts">{index ? formatSigned(index.change) : '—'}</span>
          <span className="change-pct">{index ? formatPct(index.percentChange) : '—'}</span>
        </span>

        {lastUpdated && (
          <span className="update-time">
            {lastUpdated.toLocaleTimeString('en-IN', {
              hour: '2-digit', minute: '2-digit', second: '2-digit',
              hour12: false, timeZone: 'Asia/Kolkata'
            })} IST
          </span>
        )}
      </div>

      {index && (
        <div className="header-ohlc">
          <div className="ohlc-item">
            <span className="ohlc-label">PREV</span>
            <span className="ohlc-val">{formatPoints(index.previousClose)}</span>
          </div>
          <div className="ohlc-item">
            <span className="ohlc-label">OPEN</span>
            <span className="ohlc-val">{formatPoints(index.open)}</span>
          </div>
          <div className="ohlc-item">
            <span className="ohlc-label">HIGH</span>
            <span className="ohlc-val up">{formatPoints(index.high)}</span>
          </div>
          <div className="ohlc-item">
            <span className="ohlc-label">LOW</span>
            <span className="ohlc-val down">{formatPoints(index.low)}</span>
          </div>
        </div>
      )}
    </header>
  );
}
