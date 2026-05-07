import { useState, useEffect, useRef, useCallback } from 'react';

// Android APK: backend runs on device at localhost:8080
// Browser dev:  uses REACT_APP_API_URL or localhost:8080
// Production:   set REACT_APP_API_URL at build time
const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export function useSensexData() {
  const [data, setData]               = useState(null);
  const [status, setStatus]           = useState('connecting');
  const [lastUpdated, setLastUpdated] = useState(null);
  const esRef        = useRef(null);
  const retryRef     = useRef(null);
  const retryCount   = useRef(0);
  const pollingRef   = useRef(null);

  // REST polling — fallback if SSE fails 3 times
  const startPolling = useCallback(() => {
    if (pollingRef.current) clearInterval(pollingRef.current);
    const poll = async () => {
      try {
        const res = await fetch(`${BASE_URL}/api/sensex`, { cache: 'no-store' });
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const snap = await res.json();
        setData(snap);
        setLastUpdated(new Date());
        setStatus('live');
        retryCount.current = 0;
      } catch {
        setStatus('error');
      }
    };
    poll();
    pollingRef.current = setInterval(poll, 1000);
  }, []);

  // SSE connection with exponential backoff
  const connect = useCallback(() => {
    if (esRef.current) { esRef.current.close(); esRef.current = null; }
    setStatus('connecting');

    if (typeof EventSource === 'undefined') { startPolling(); return; }

    try {
      const es = new EventSource(`${BASE_URL}/api/sensex/stream`);
      esRef.current = es;

      es.addEventListener('sensex', (e) => {
        try {
          const snap = JSON.parse(e.data);
          setData(snap);
          setLastUpdated(new Date());
          setStatus('live');
          retryCount.current = 0;
        } catch (err) {
          console.warn('[sse] parse error', err);
        }
      });

      es.addEventListener('waiting', () => setStatus('connecting'));

      es.onerror = () => {
        es.close(); esRef.current = null;
        setStatus('error');
        if (retryCount.current >= 3) { startPolling(); return; }
        const delay = Math.min(1000 * Math.pow(2, retryCount.current), 4000);
        retryCount.current += 1;
        retryRef.current = setTimeout(connect, delay);
      };
    } catch (err) {
      console.warn('[sse] init failed, polling', err);
      startPolling();
    }
  }, [startPolling]);

  useEffect(() => {
    connect();
    return () => {
      if (esRef.current)    { esRef.current.close(); esRef.current = null; }
      if (retryRef.current)  clearTimeout(retryRef.current);
      if (pollingRef.current) clearInterval(pollingRef.current);
    };
  }, [connect]);

  return { data, status, lastUpdated };
}
