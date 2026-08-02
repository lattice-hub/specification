cmake_minimum_required(VERSION 3.20)

project(pole_specification_cpp LANGUAGES CXX)

set(POLE_SPECIFICATION_VERSION "0.1.0" CACHE STRING "Pole Specification package version")

find_package(Protobuf 5.26.1 CONFIG REQUIRED)
find_package(gRPC 1.64 CONFIG REQUIRED)
find_package(Python3 REQUIRED COMPONENTS Interpreter)

if(TARGET protobuf::protoc)
    set(POLE_PROTOC "$<TARGET_FILE:protobuf::protoc>")
else()
    set(POLE_PROTOC "${Protobuf_PROTOC_EXECUTABLE}")
endif()

set(POLE_API_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/../../../api")
set(POLE_GENERATED_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}/generated")
set(POLE_PROTO_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}/proto")

file(GLOB_RECURSE POLE_PROTO_FILES CONFIGURE_DEPENDS "${POLE_API_DIRECTORY}/*.proto")
if(NOT POLE_PROTO_FILES)
    message(FATAL_ERROR "No authoritative Proto files found under ${POLE_API_DIRECTORY}")
endif()

set(POLE_GENERATED_SOURCES)
set(POLE_GENERATED_HEADERS)
foreach(POLE_PROTO_FILE IN LISTS POLE_PROTO_FILES)
    get_filename_component(POLE_PROTO_NAME "${POLE_PROTO_FILE}" NAME_WE)
    list(APPEND POLE_GENERATED_SOURCES
        "${POLE_GENERATED_DIRECTORY}/${POLE_PROTO_NAME}.pb.cc"
        "${POLE_GENERATED_DIRECTORY}/${POLE_PROTO_NAME}.grpc.pb.cc"
    )
    list(APPEND POLE_GENERATED_HEADERS
        "${POLE_GENERATED_DIRECTORY}/${POLE_PROTO_NAME}.pb.h"
        "${POLE_GENERATED_DIRECTORY}/${POLE_PROTO_NAME}.grpc.pb.h"
    )
endforeach()

set(POLE_PROTOBUF_INCLUDE_ARGUMENTS)
foreach(POLE_PROTOBUF_INCLUDE_DIRECTORY IN LISTS Protobuf_INCLUDE_DIRS)
    list(APPEND POLE_PROTOBUF_INCLUDE_ARGUMENTS
        --include-dir "${POLE_PROTOBUF_INCLUDE_DIRECTORY}"
    )
endforeach()

add_custom_command(
    OUTPUT
        ${POLE_GENERATED_SOURCES}
        ${POLE_GENERATED_HEADERS}
        "${POLE_GENERATED_DIRECTORY}/proto-files.txt"
    COMMAND
        "${Python3_EXECUTABLE}"
        "${CMAKE_CURRENT_SOURCE_DIR}/scripts/generate.py"
        --api-dir "${POLE_API_DIRECTORY}"
        --output-dir "${POLE_GENERATED_DIRECTORY}"
        --proto-dir "${POLE_PROTO_DIRECTORY}"
        --protoc "${POLE_PROTOC}"
        --grpc-cpp-plugin "$<TARGET_FILE:gRPC::grpc_cpp_plugin>"
        ${POLE_PROTOBUF_INCLUDE_ARGUMENTS}
    DEPENDS
        ${POLE_PROTO_FILES}
        "${CMAKE_CURRENT_SOURCE_DIR}/scripts/generate.py"
        protobuf::protoc
        gRPC::grpc_cpp_plugin
    VERBATIM
)

add_library(pole-specification ${POLE_GENERATED_SOURCES} ${POLE_GENERATED_HEADERS})
add_library(PoleSpecification::pole-specification ALIAS pole-specification)

target_compile_features(pole-specification PUBLIC cxx_std_17)
target_include_directories(
    pole-specification
    PUBLIC
        $<BUILD_INTERFACE:${POLE_GENERATED_DIRECTORY}>
        $<INSTALL_INTERFACE:include/pole/specification>
)
target_link_libraries(pole-specification PUBLIC protobuf::libprotobuf gRPC::grpc++)

include(CTest)
if(BUILD_TESTING)
    add_executable(
        pole-specification-generated-api-test
        tests/generated_api_test.cc
    )
    target_link_libraries(
        pole-specification-generated-api-test
        PRIVATE PoleSpecification::pole-specification
    )
    add_test(
        NAME pole-specification-generated-api
        COMMAND pole-specification-generated-api-test
    )
    add_test(
        NAME pole-specification-generated-completeness
        COMMAND
            "${Python3_EXECUTABLE}"
            "${CMAKE_CURRENT_SOURCE_DIR}/tests/verify_generated.py"
            --api-dir "${POLE_API_DIRECTORY}"
            --generated-dir "${POLE_GENERATED_DIRECTORY}"
            --proto-dir "${POLE_PROTO_DIRECTORY}"
    )
endif()

include(CMakePackageConfigHelpers)
configure_package_config_file(
    cmake/PoleSpecificationConfig.cmake.in
    "${CMAKE_CURRENT_BINARY_DIR}/PoleSpecificationConfig.cmake"
    INSTALL_DESTINATION lib/cmake/PoleSpecification
)
write_basic_package_version_file(
    "${CMAKE_CURRENT_BINARY_DIR}/PoleSpecificationConfigVersion.cmake"
    VERSION 0.1.0
    COMPATIBILITY SameMajorVersion
)

install(
    TARGETS pole-specification
    EXPORT PoleSpecificationTargets
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin
)
install(FILES ${POLE_GENERATED_HEADERS} DESTINATION include/pole/specification)
install(
    EXPORT PoleSpecificationTargets
    FILE PoleSpecificationTargets.cmake
    NAMESPACE PoleSpecification::
    DESTINATION lib/cmake/PoleSpecification
)
install(
    FILES
        "${CMAKE_CURRENT_BINARY_DIR}/PoleSpecificationConfig.cmake"
        "${CMAKE_CURRENT_BINARY_DIR}/PoleSpecificationConfigVersion.cmake"
    DESTINATION lib/cmake/PoleSpecification
)
