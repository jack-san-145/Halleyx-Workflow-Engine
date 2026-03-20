# Halleyx Workflow Engine

Halleyx is a simple workflow engine built with Go.

The idea is to define workflows as a series of steps and let the system execute them automatically. Each step runs inside a Docker container, which keeps execution isolated and flexible.

This project reflects the architecture of real-world workflow engines, built to handle scalable execution, rule-based orchestration, and high-performance task processing.

---

## What this project does

- Create workflows  
- Add steps to workflows  
- Define rules to control execution flow  
- Execute workflows step by step  
- Run task steps inside Docker containers  
- Track execution in real time  

---

## Core concept

A workflow is made up of steps.

Each step can be:
- task  
- approval  
- notification  

Rules decide what happens after each step.

When a workflow starts:
1. Execution is created  
2. First step runs  
3. Rules are evaluated  
4. Next step is selected  
5. Process continues until completion  

---

## Docker based execution

Task steps run inside Docker containers.

This gives:
- Isolated execution  
- No dependency issues  
- Safe runtime environment  
- Ability to run any tool (gcc, python, node, etc.)  

---

## High performance backend

The backend is built using Golang, designed from the ground up for high concurrency and efficient resource usage.

- Each user request is handled in a separate goroutine  
- Extremely lightweight (consuming only a few KB of memory per routine)  
- Non-blocking architecture ensures minimal latency  
- Efficient CPU usage even under heavy load  

Because of this architecture:

- The system can handle thousands of concurrent users without degradation
- Memory usage stays very low (KB level instead of MB per request)  
- Performs reliably even on low-end infrastructure  
- Scales well under high traffic  

This makes the backend highly resilient, efficient, and capable of sustaining heavy traffic loads without performance bottlenecks.

---

## Example: C Program Pipeline

This workflow compiles and runs a C program.

---

### Input Schema

```json
{
  "code": {
    "description": "Source code to compile and run",
    "required": true,
    "type": "string"
  }
}
```

---

### Step 1: Compile

```json
{
  "command": [
    "sh",
    "-c",
    "gcc main.c -o app; echo $? > /workspace/exitcode"
  ],
  "image": "gcc:latest",
  "timeout": 60,
  "volumes": [
    "/tmp/exec-{{execution_id}}:/workspace"
  ],
  "workdir": "/workspace"
}
```

#### Rules

- Condition: `exit_code == 0`  
  → Next: run step  
  Priority: 1  

- Condition: `DEFAULT`  
  → End  
  Priority: 100  

---

### Step 2: Run

```json
{
  "command": [
    "sh",
    "-c",
    "/workspace/app"
  ],
  "image": "ubuntu:latest",
  "volumes": [
    "/tmp/exec-{{execution_id}}:/workspace"
  ],
  "workdir": "/workspace"
}
```

#### Rules

- Condition: `DEFAULT`  
  → End  
  Priority: 100  

---

### Example C Code

```c
#include <stdio.h>

int main() {
    printf("Hello, World!\n");
    return 0;
}
```

---

## Project structure

```
cmd/server          entry point  
frontend            UI  
internal/api        handlers  
internal/engine     execution logic  
internal/store      database layer  
internal/ws         websocket hub  
migrations          SQL schema  
```

---

## Tech stack

- Go  
- PostgreSQL  
- Docker  
- WebSockets  
- HTML, CSS, JavaScript  

---

## Setup

### 1. Create `.env` file

```env
POSTGRES_DB=halleyx
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

DATABASE_URL=postgres://postgres:postgres@postgres:5432/halleyx?sslmode=disable
```

---

### 2. Start application

```bash
docker compose up --build
```

---

### 3. Run migrations

```bash
docker compose exec postgres psql -U postgres -d halleyx -f /migrations/001_init.sql
```

---

### 4. Open in browser

```
http://localhost:8080
```

---

## Notes

- Postgres may take a few seconds to start  
- Backend depends on DB connection  
- Docker socket is used for running containers  

---

## What this project demonstrates

- Workflow based system design  
- Rule driven execution  
- Docker based task execution  
- Real time updates using WebSockets  
- Backend architecture design  

---

## Future improvements

- Authentication  
- Role based approvals  
- Retry logic  
- Better UI  
- Logging and monitoring  

---
