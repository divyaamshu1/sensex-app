package com.sensexlive.app;

import android.os.Bundle;
import android.util.Log;
import com.getcapacitor.BridgeActivity;

/**
 * Main entry point for the SENSEX Live Android app.
 *
 * Extends BridgeActivity (Capacitor) so the React WebView works normally.
 * Additionally auto-starts the embedded Go backend before the WebView loads,
 * so no Termux / manual steps are ever needed.
 */
public class MainActivity extends BridgeActivity {

    private static final String TAG = "SensexMain";
    private BackendService backend;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        // 1. Start the Go backend first (extracts binary from assets + launches it)
        backend = new BackendService(this);
        backend.start();

        // 2. Wait up to 5 s for the server to be reachable before loading the WebView.
        //    This prevents the "connection refused" flash on cold start.
        new Thread(() -> {
            boolean ready = backend.waitUntilReady(5000);
            if (ready) {
                Log.i(TAG, "Backend ready — WebView will connect successfully");
            } else {
                Log.w(TAG, "Backend not ready after 5 s — WebView may show retry");
            }
        }).start();

        // 3. Init Capacitor bridge + load React app normally
        super.onCreate(savedInstanceState);
    }

    @Override
    public void onDestroy() {
        super.onDestroy();
        if (backend != null) {
            backend.stop();
        }
    }
}
