# SENSEX Live — Full-Stack App

Live BSE Sensex tracker. Go backend fetches index + constituent weightages from BSE every second. React frontend shows each stock's point contribution in real time.

**Stack**: Go (backend, embedded in APK) · React + Capacitor (Android APK) · GitHub Actions (CI/CD)

---

## How it works — fully automatic, zero Termux needed

```
User installs APK
      ↓
Opens SENSEX Live
      ↓
MainActivity.onCreate()
  → BackendService.start()
    → extracts Go binary from APK assets to app's private dir
    → chmod +x, launches ./sensex-server on :8080
    → watchdog thread monitors it (auto-restarts if it crashes)
      ↓
React WebView loads → hits localhost:8080 ✅
```

The Go backend is bundled **inside** the APK as an asset (`sensex-server`).  
On every launch the app extracts it, makes it executable, and starts it automatically.  
No Termux. No manual steps. Install → open → live data.

---

## Repo structure

```
sensex-app/
├── backend/                         ← Go server (stdlib only, zero deps)
│   ├── cmd/server/main.go
│   └── internal/
│       ├── api/handler.go           ← REST + SSE endpoints
│       ├── cache/cache.go           ← Thread-safe snapshot store
│       ├── fetcher/fetcher.go       ← BSE HTTP fetchers + fallback
│       ├── fetcher/poller.go        ← Background goroutines
│       └── models/models.go
│
├── frontend/                        ← React app (Capacitor wrapper)
│   ├── src/
│   │   ├── App.js / App.css
│   │   ├── components/
│   │   ├── hooks/useSensexData.js   ← SSE + polling fallback
│   │   └── utils/format.js
│   ├── public/index.html
│   ├── capacitor.config.json
│   └── package.json
│
├── android/                         ← Android Java sources (copied into
│   └── app/src/main/                  Capacitor project by CI)
│       ├── java/com/sensexlive/app/
│       │   ├── BackendService.java  ← Manages Go binary lifecycle
│       │   └── MainActivity.java   ← Entry point, starts backend
│       ├── AndroidManifest.xml
│       └── res/
│           ├── xml/network_security_config.xml
│           └── values/{strings,colors,styles}.xml
│
└── .github/workflows/build.yml      ← CI: push → APK (with Go binary embedded)
```

---

## Get the APK in 4 steps (no PC needed)

### Step 1 — Create a GitHub repo

1. Open **github.com** in your phone browser
2. Tap **+** → **New repository** → name it `sensex-app` → **Create repository**

### Step 2 — Upload the code

**Option A — GitHub web UI (easiest)**
1. In your new repo, tap **uploading an existing file**
2. Drag all folders/files from this zip
3. Commit with message `initial commit`

**Option B — Termux**
```bash
pkg install git
git clone https://github.com/YOUR_USERNAME/sensex-app
# copy files in, then:
git add . && git commit -m "initial" && git push
```

### Step 3 — GitHub Actions builds automatically

After you push, GitHub starts building.

- Go to your repo → **Actions** tab
- Watch **Build Sensex APK** run (≈ 8–12 min)
- Green checkmark = done ✅

### Step 4 — Download and install

1. Click the completed workflow run
2. Scroll to **Artifacts** → tap **SensexLive-APK**
3. Open the downloaded `SensexLive-debug.apk`
4. Enable **Install from unknown sources** if prompted → Install
5. Open the app — backend starts automatically 🚀

---

## CI pipeline overview

```
push to main
    │
    ├─ Job 1: Go Backend
    │   ├─ Build Linux amd64
    │   ├─ Build Android ARM64   ← embedded in APK
    │   └─ Build Android ARM
    │
    └─ Job 2: Android APK  (needs: build-backend)
        ├─ Download ARM64 binary
        ├─ Build React app
        ├─ Capacitor sync android
        ├─ Copy binary → android/app/src/main/assets/sensex-server  ← KEY STEP
        ├─ Copy BackendService.java + MainActivity.java into project
        ├─ Copy AndroidManifest + res overrides
        ├─ Patch build.gradle (noCompress for binary)
        └─ ./gradlew assembleDebug → SensexLive-debug.apk
```

---

## API endpoints

| Method | URL | Description |
|--------|-----|-------------|
| GET | `http://localhost:8080/api/sensex` | Latest snapshot (JSON) |
| GET | `http://localhost:8080/api/sensex/stream` | Live SSE stream |
| GET | `http://localhost:8080/api/health` | Health check |

---

## Local development (needs a PC)

```bash
# Terminal 1 — Go backend
cd backend
go run ./cmd/server

# Terminal 2 — React frontend
cd frontend
npm install
npm start
# Opens http://localhost:3000
```

---

## Data sources

| Data | Source | Refresh |
|------|--------|---------|
| Sensex index price | `api.bseindia.com/BseIndiaAPI/api/SensexData/w` | Every 1s |
| Constituent weightages | Same API with `?flag=16` | Every 5 min |
| Fallback index | NSE mirror | On primary failure |

> Market hours: **Mon–Fri 9:15 AM – 3:30 PM IST**. Prices don't change outside these hours.
