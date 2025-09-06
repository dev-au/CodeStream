# CodeStream

CodeStream is an open-source collaborative coding interview platform.
It allows interviewers and candidates to code together in real time, run solutions securely, and manage interview
sessions seamlessly.

---

## 🚀 Quick Start

### 1. Clone the repository

git clone [https://github.com/dev-au/CodeStream.git](https://github.com/dev-au/CodeStream.git)
cd CodeStream


### 2. 🛠 Build Runner Images

Each language has its own Dockerfile under `runner-images/<lang>`.

```bash
# All runner
docker build -f runner.Dockerfile -t runner-code:latest .
```

### 3. Build and run with Docker Compose

`docker-compose up --build`

This will start the full platform including backend, frontend, and runner services.

### 4. Access the platform

Open your browser at:
[http://localhost:8000](http://localhost:8000)

---
