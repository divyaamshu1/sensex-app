// Format a number as Indian currency style (e.g. 73,456.78)
export function formatPoints(n, decimals = 2) {
  if (n == null || isNaN(n)) return '—';
  return n.toLocaleString('en-IN', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

// Format a signed number with + or - prefix
export function formatSigned(n, decimals = 2) {
  if (n == null || isNaN(n)) return '—';
  const abs = Math.abs(n).toLocaleString('en-IN', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
  return (n >= 0 ? '+' : '−') + abs;
}

// Format percent
export function formatPct(n) {
  if (n == null || isNaN(n)) return '—';
  const abs = Math.abs(n).toFixed(2);
  return (n >= 0 ? '+' : '−') + abs + '%';
}

// Format time as HH:MM:SS IST
export function formatTime(date) {
  if (!date) return '—';
  return date.toLocaleTimeString('en-IN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
    timeZone: 'Asia/Kolkata',
  });
}

// Returns 'up', 'down', or 'flat'
export function direction(n) {
  if (n > 0) return 'up';
  if (n < 0) return 'down';
  return 'flat';
}
