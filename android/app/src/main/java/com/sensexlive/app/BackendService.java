package com.sensexlive.app;

import android.content.Context;
import android.util.Log;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;

/**
 * BackendService manages the embedded Go server binary.
 *
 * Lifecycle:
 *  1. On first launch, copies the binary from APK assets → app's private files dir
 *  2. Sets executable permission
 *  3. Starts the process on port 8080
 *  4. Monitors it in a background thread; restarts if it dies
 *  5. Stops cleanly when the app is destroyed
 */
public class BackendService {

    private static final String TAG         = "SensexBackend";
    private static final String ASSET_NAME  = "sensex-server";   // file in assets/
    private static final String BIN_NAME    = "sensex-server";   // name on disk
    private static final int    PORT        = 8080;
    private static final int    MAX_RETRIES = 10;

    private final Context context;
    private Process  process;
    private Thread   watchdog;
    private volatile boolean stopped = false;

    public BackendService(Context context) {
        this.context = context.getApplicationContext();
    }

    // ── Public API ─────────────────────────────────────────────────

    /** Start the backend. Safe to call multiple times. */
    public synchronized void start() {
        if (isRunning()) {
            Log.i(TAG, "Backend already running");
            return;
        }
        stopped = false;
        try {
            File binary = extractBinary();
            launchProcess(binary);
            startWatchdog();
        } catch (Exception e) {
            Log.e(TAG, "Failed to start backend", e);
        }
    }

    /** Stop the backend and watchdog. */
    public synchronized void stop() {
        stopped = true;
        if (process != null) {
            process.destroy();
            process = null;
        }
        if (watchdog != null) {
            watchdog.interrupt();
            watchdog = null;
        }
        Log.i(TAG, "Backend stopped");
    }

    /** True if the Go process is alive. */
    public boolean isRunning() {
        if (process == null) return false;
        try {
            process.exitValue();
            return false; // exited
        } catch (IllegalThreadStateException e) {
            return true;  // still running
        }
    }

    /**
     * Block until the HTTP health endpoint responds OK, or timeout.
     * Returns true if ready within timeoutMs.
     */
    public boolean waitUntilReady(int timeoutMs) {
        long deadline = System.currentTimeMillis() + timeoutMs;
        while (System.currentTimeMillis() < deadline) {
            if (pingHealth()) return true;
            try { Thread.sleep(200); } catch (InterruptedException ignored) {}
        }
        return false;
    }

    // ── Private helpers ────────────────────────────────────────────

    /**
     * Copy the Go binary from APK assets to the app's private files dir.
     * Always re-extracts so updates are picked up when APK changes.
     */
    private File extractBinary() throws IOException {
        File binDir  = new File(context.getFilesDir(), "bin");
        File binFile = new File(binDir, BIN_NAME);

        if (!binDir.exists()) binDir.mkdirs();

        try (InputStream in  = context.getAssets().open(ASSET_NAME);
             OutputStream out = new FileOutputStream(binFile)) {
            byte[] buf = new byte[65536];
            int n;
            while ((n = in.read(buf)) != -1) out.write(buf, 0, n);
        }

        // chmod +x
        binFile.setExecutable(true, true);
        Log.i(TAG, "Binary extracted to: " + binFile.getAbsolutePath()
                + " (" + binFile.length() + " bytes)");
        return binFile;
    }

    /** Launch the Go binary as a child process. */
    private void launchProcess(File binary) throws IOException {
        ProcessBuilder pb = new ProcessBuilder(binary.getAbsolutePath());
        pb.environment().put("PORT", String.valueOf(PORT));
        pb.redirectErrorStream(true);
        process = pb.start();
        Log.i(TAG, "Go backend process started, port=" + PORT);

        // Drain stdout → logcat so we can see Go logs
        final Process p = process;
        Thread logger = new Thread(() -> {
            try (java.io.BufferedReader r = new java.io.BufferedReader(
                    new java.io.InputStreamReader(p.getInputStream()))) {
                String line;
                while ((line = r.readLine()) != null) {
                    Log.d(TAG, "[go] " + line);
                }
            } catch (IOException ignored) {}
        }, "go-logger");
        logger.setDaemon(true);
        logger.start();
    }

    /**
     * Watchdog: waits for the process to exit, then restarts it
     * unless stop() was called. Uses exponential backoff.
     */
    private void startWatchdog() {
        watchdog = new Thread(() -> {
            int retries = 0;
            while (!stopped) {
                try {
                    if (process != null) {
                        int code = process.waitFor();
                        if (stopped) break;
                        Log.w(TAG, "Go backend exited with code " + code + ", restarting...");
                    }
                } catch (InterruptedException e) {
                    break;
                }

                if (retries >= MAX_RETRIES) {
                    Log.e(TAG, "Too many restarts, giving up");
                    break;
                }

                long delay = Math.min(1000L * (1L << retries), 16000L); // 1s 2s 4s … 16s
                retries++;
                Log.i(TAG, "Restart #" + retries + " in " + delay + "ms");

                try { Thread.sleep(delay); } catch (InterruptedException e) { break; }

                if (!stopped) {
                    try {
                        File binary = extractBinary();
                        launchProcess(binary);
                        retries = 0; // reset after successful launch
                    } catch (IOException e) {
                        Log.e(TAG, "Restart failed", e);
                    }
                }
            }
        }, "go-watchdog");
        watchdog.setDaemon(true);
        watchdog.start();
    }

    /** Quick HTTP GET to /api/health to check if the server is up. */
    private boolean pingHealth() {
        try {
            URL url = new URL("http://localhost:" + PORT + "/api/health");
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setConnectTimeout(500);
            conn.setReadTimeout(500);
            conn.setRequestMethod("GET");
            int code = conn.getResponseCode();
            conn.disconnect();
            return code == 200;
        } catch (Exception e) {
            return false;
        }
    }
}
