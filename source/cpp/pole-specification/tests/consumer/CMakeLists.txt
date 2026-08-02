cmake_minimum_required(VERSION 3.20)

project(pole_specification_consumer LANGUAGES CXX)

find_package(PoleSpecification CONFIG REQUIRED)

add_executable(
    pole-specification-consumer
    ../generated_api_test.cc
)
target_link_libraries(
    pole-specification-consumer
    PRIVATE PoleSpecification::pole-specification
)
