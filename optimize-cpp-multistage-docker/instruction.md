The C++ application located in `/app/src` is currently failing to compile due to a configuration issue, and its corresponding `/app/src/Dockerfile` produces a bloated, insecure image.

The codebase needs to be updated so that the application compiles successfully. Furthermore, the Docker containerization must be optimized for production. The final built Docker image must satisfy the following constraints:
- It uses the canonical `debian:bookworm-slim@sha256:4724b8cc51e33e398f0e2e15e18d5ec2851ff0c2280647e1310bc1642182655d` image as its base.
- It is a minimal runtime environment that does not contain build toolchains (such as `g++` or `cmake`).
- It executes the application safely as a non-root user named `appuser`.
- The successfully compiled binary is accessible at `/usr/local/bin/app` within the final image.
