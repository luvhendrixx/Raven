### Project Layout/Structure

```text
raven/
├── cmd/
│   └── raven-server/
│       └── main.go
├── internal/
│   ├── network/
│   ├── protocol/
│   └── server/
├── pkg/
├── bindings/
├── services/
├── web/
├── apps/
├── docs/
├── go.mod
├── go.sum
└── README.md

```

---

### How Each Directory Works

**1. `cmd/` (Application Entry Points)**

* **Purpose:** Holds the main executable entry points for your Go project.
* **Role:** Inside `cmd/raven-server/main.go`, you write minimal initialization code: reading configs, setting up logging, and calling your actual server logic from `internal/`. Keep logic out of `cmd/`—it should strictly be the trigger that boots the application.

**2. `internal/` (Private Go Architecture)**

* **Purpose:** Go’s compiler strictly enforces that anything inside `internal/` cannot be imported by external Go modules.
* **Role:** This is where 90% of your Go backend code will live. As you build the networking layer brick by brick, you can break it down into clean sub-packages:
* `internal/network/`: Connection handling, socket listeners, and connection pools.
* `internal/protocol/`: Packet parsing, message framing, and serialization logic.
* `internal/server/`: Core application state, client session routing, and event dispatching.



**3. `pkg/` (Public Go Libraries - Optional)**

* **Purpose:** Code you *want* other Go projects or external callers to import.
* **Role:** If you eventually build a Go-based SDK, CLI tool, or shared protocol package that other external programs should consume, place it here. If everything is private to Raven, leave this empty or skip it for now.

**4. `bindings/` (Cross-Language Interop)**

* **Purpose:** Bridging Go with low-level native languages (C, C++, Rust).
* **Role:** If you later write a performance-critical packet processor in Rust or C/C++ and interface with Go via CGO or Foreign Function Interfaces (FFI), keep the native code, compiled headers, and wrapper glue code inside this directory to prevent low-level files from polluting your Go packages.

**5. `services/` (Microservices & Secondary Backends)**

* **Purpose:** Homes for separate backend services written in other languages (e.g., Python, Rust).
* **Role:** If you create a specialized Python service for data processing, AI, or automated bots, or a standalone Rust microservice, give each its own isolated folder inside `services/` (e.g., `services/analytics-py/`, `services/router-rs/`).

**6. `web/` and `apps/` (Frontends)**

* **Purpose:** Isolating user interfaces from the backend systems.
* **Role:** Use `web/` for browser clients (React, Vue, Svelte, or plain JS). Use `apps/` if you later add desktop (Tauri/Electron) or mobile applications (Flutter/React Native).

---

### Key Architectural Guidelines

* **Start Monolithic inside a Single Repository:** Keep all components in one repository (a monorepo layout). It makes cross-language development, protocol updates, and local testing drastically easier while learning.
* **Define clear network boundaries early:** Since you might use multiple languages, avoid tightly coupling your Go code to a specific language runtime. Communicate across components using clear protocols like WebSockets, gRPC, raw TCP with binary framing, or JSON/Protobuf over streams.
* **Keep Go package names focused:** In Go, package names should be concise, lowercase, single-word nouns describing their purpose (e.g., `package network`, `package protocol`). Avoid generic folder names like `helpers`, `utils`, or `common`.
*  When you create many empty folders upfront (like bindings/, services/, web/, protocol/), it creates visual clutter in your editor tree view.

Cause as a single developer, starting with a minimal folder structure allows you to focus only on what exists and works today. You can always run `mkdir -p` to create a directory when you are ready to write code for it.
