cmake_minimum_required(VERSION 3.16)
project(bustamante_print_tools CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

find_package(Qt6 REQUIRED COMPONENTS Core Widgets)

# Windows: link the setupapi library for printer enumeration.
if(WIN32)
    set(PLATFORM_LIBS setupapi)
endif()

set(SOURCES
    src/main.cpp
    src/mainwindow.cpp
    src/utils.cpp
    src/status.cpp
    src/devices.cpp
    src/transport.cpp
    src/d4.cpp
    src/printer.cpp
)

set(HEADERS
    include/mainwindow.h
    include/utils.h
    include/status.h
    include/devices.h
    include/transport.h
    include/d4.h
    include/printer.h
)

add_executable(bustamante_print_tools ${SOURCES} ${HEADERS})

target_include_directories(bustamante_print_tools PRIVATE include)

target_link_libraries(bustamante_print_tools PRIVATE
    Qt6::Core
    Qt6::Widgets
    ${PLATFORM_LIBS}
)

# Copy devices.xml next to the binary so the app can find it.
add_custom_command(TARGET bustamante_print_tools POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy_if_different
        ${CMAKE_CURRENT_SOURCE_DIR}/devices.xml
        $<TARGET_FILE_DIR:bustamante_print_tools>/devices.xml)
