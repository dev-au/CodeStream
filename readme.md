# CodeStream

CodeStream is an open-source collaborative coding interview platform.
It allows interviewers and candidates to code together in real time, run solutions securely, and manage interview
sessions seamlessly.

---

## 🚀 Quick Start

### 1. Clone the repository

git clone [https://github.com/dev-au/CodeStream.git](https://github.com/dev-au/CodeStream.git)
cd CodeStream

### 2. Build and run with Docker Compose

docker-compose up --build

This will start the full platform including backend, frontend, and runner services.

### 3. Access the platform

Open your browser at:
[http://localhost:8000](http://localhost:8000)

---

## 🛠 Build Runner Images

Each language has its own Dockerfile under `runner-images/<lang>`.
For example:

```bash
# Python runner
docker build -t runner-python:latest runner-images/runner-python

# JavaScript runner
docker build -t runner-node:latest runner-images/runner-node
```
---

## ⚙️ Tech Stack

* Backend: Go (Gin, WebSocket, Redis)
* Execution Layer: Docker runner images
* Storage: Redis

