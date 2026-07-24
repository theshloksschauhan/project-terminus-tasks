"""Tests for C++ CMake fix and Dockerfile multi-stage optimization."""

import subprocess
from pathlib import Path
from dockerfile_parse import DockerfileParser

def test_cmake_compiles_successfully():
    """Verify that the C++ code compiles successfully with the fixed CMakeLists.txt."""
    src_dir = Path("/app/src")
    build_dir = src_dir / "build"
    build_dir.mkdir(exist_ok=True)
    
    # Run cmake
    cmake_result = subprocess.run(["cmake", ".."], cwd=build_dir, capture_output=True, text=True)
    assert cmake_result.returncode == 0, f"CMake failed: {cmake_result.stderr}"
    
    # Run make
    make_result = subprocess.run(["make"], cwd=build_dir, capture_output=True, text=True)
    assert make_result.returncode == 0, f"Make failed: {make_result.stderr}"
    
    # Run binary
    app_path = build_dir / "app"
    assert app_path.exists(), "Binary 'app' was not created."
    app_result = subprocess.run([str(app_path)], capture_output=True, text=True)
    assert app_result.returncode == 0
    assert "Terminus C++ Build Successful!" in app_result.stdout

def test_dockerfile_is_multistage():
    """Verify the Dockerfile uses a multi-stage build."""
    df_path = "/app/src/Dockerfile"
    with open(df_path, "r") as f:
        content = f.read()
    
    from_count = content.count("FROM ")
    assert from_count >= 2, "Dockerfile must use multi-stage builds (multiple FROM statements)."

def test_dockerfile_final_base_image():
    """Verify the final stage uses the specified digest-pinned base image."""
    parser = DockerfileParser("/app/src")
    
    from_instructions = [ins for ins in parser.structure if ins['instruction'] == 'FROM']
    assert len(from_instructions) >= 2, "Need at least two stages."
    
    final_from = from_instructions[-1]['value']
    expected_image = "debian:bookworm-slim@sha256:4724b8cc51e33e398f0e2e15e18d5ec2851ff0c2280647e1310bc1642182655d"
    
    image_part = final_from.split()[0]
    assert image_part == expected_image, f"Final stage base image must be {expected_image}"

def test_dockerfile_non_root_user():
    """Verify the final runtime image runs as non-root 'appuser'."""
    parser = DockerfileParser("/app/src")
    
    user_instructions = [ins for ins in parser.structure if ins['instruction'] == 'USER']
    assert len(user_instructions) > 0, "Dockerfile is missing a USER directive."
    
    final_user = user_instructions[-1]['value']
    assert final_user == "appuser", "Final runtime image must run as 'appuser'."

def test_dockerfile_no_compilers_in_final_stage():
    """Verify the final stage does not install cmake or g++ and correctly copies the binary."""
    parser = DockerfileParser("/app/src")
    
    last_from_idx = -1
    for i, ins in enumerate(parser.structure):
        if ins['instruction'] == 'FROM':
            last_from_idx = i
            
    final_stage_instructions = parser.structure[last_from_idx:]
    
    for ins in final_stage_instructions:
        if ins['instruction'] == 'RUN':
            cmd = ins['value'].lower()
            assert "g++" not in cmd and "cmake" not in cmd, "Final stage must not install build tools like g++ or cmake."
            
        if ins['instruction'] == 'COPY':
            val = ins['value']
            assert "--from=" in val, "Final stage should only COPY artifacts from the builder stage, not local context."
            assert "/usr/local/bin/app" in val, "Compiled binary should be placed in /usr/local/bin/app."
