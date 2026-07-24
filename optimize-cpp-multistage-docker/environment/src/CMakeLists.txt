cmake_minimum_required(VERSION 3.10)
project(optimize_cpp)

set(CMAKE_CXX_STANDARD 17)

add_library(mylib lib.cpp)
add_executable(app main.cpp)
# BUG: missing target_link_libraries(app mylib)
