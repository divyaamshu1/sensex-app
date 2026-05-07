import React from 'react';
import { useSensexData } from './hooks/useSensexData';
import SensexHeader from './components/SensexHeader';
import ConstituentsList from './components/ConstituentsList';
import './App.css';

export default function App() {
  const { data, status, lastUpdated } = useSensexData();

  return (
    <div className="app">
      <div className="app-inner">
        <SensexHeader
          index={data?.index}
          status={status}
          lastUpdated={lastUpdated}
        />

        {data?.error && (
          <div className="error-banner">
            ⚠ {data.error}
          </div>
        )}

        <ConstituentsList
          constituents={data?.constituents}
          sensexChange={data?.index?.change}
        />

        <footer className="app-footer">
          <span>Data: BSE India · Updates every 1s</span>
          <span>Market hrs 9:15–15:30 IST</span>
        </footer>
      </div>
    </div>
  );
}
