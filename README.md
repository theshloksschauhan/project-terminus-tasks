# Project Terminus Tasks

This repository contains two complex engineering tasks I built for Snorkel's **Project Terminus**, a benchmark for evaluating frontier AI coding agents on real-world, multi-step engineering challenges.

Project Terminus goes beyond simple code completion. It tests agents in fully containerized environments on tasks that demand deep reasoning, debugging, and environmental interaction.

## Tasks Included

### 1. `optimize-cpp-multistage-docker`
**Category:** Build and Dependency Management  
**Languages:** C++, Docker, Bash  
**Difficulty:** Hard  

**Overview:** 
This task challenges an agent to debug a failing C++ compilation pipeline and completely rewrite an inefficient, insecure Dockerfile into a lean, optimized multi-stage build. The agent must:
- Fix a broken `CMakeLists.txt` linking issue.
- Convert a monolithic Dockerfile into a multi-stage build.
- Implement robust security practices (dropping privileges with a non-root user).
- Ensure the final runtime image is minimal and lacks build toolchains.

### 2. `apt-mirror-journal-repair`
**Category:** System Administration  
**Languages:** Go, Bash  
**Difficulty:** Hard  

**Overview:**
This task involves repairing an offline package archive sync daemon written in Go. The daemon loses state coherence when long network pulls are interrupted. The agent is required to:
- Deeply trace a driver call graph across four separate Go modules.
- Ensure deterministic convergence of synchronization operations.
- Preserve hook-order sensitivities and robustly survive permutations in network drills.
- The task emphasizes purely behavioral validation (via `pytest`), aggressively penalizing brittle, step-by-step imitation or string matching.

## Evaluation
Both tasks are structured strictly against the Terminus Edition 2 guidelines. They are validated deterministically via an `oracle` solution and a comprehensive Python `pytest` suite running within a Docker sandbox, ensuring no agent can "cheat" its way to success.

## Task layout

Each task follows the same structure:

```
<task>/
  instruction.md      what the agent is asked to do
  task.toml           metadata, timeouts and resource limits
  environment/        Dockerfile and the broken codebase the agent starts from
  solution/solve.sh   reference (oracle) solution
  tests/test.sh       runs tests/test_outputs.py with pytest and writes reward 1 or 0
```

## Verify a task locally

```bash
cd optimize-cpp-multistage-docker
docker build -t terminus-task environment/
docker run --rm -it \
  -v "$PWD/solution:/solution" -v "$PWD/tests:/tests" \
  terminus-task bash -c "bash /solution/solve.sh && bash /tests/test.sh"
```

The oracle solution should score a reward of 1. The unmodified environment should score 0.
