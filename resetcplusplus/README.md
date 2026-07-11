# Bustamante Print Tools — C++ (Qt6)

Versión en C++ con interfaz gráfica (Qt6) de la herramienta de reset de tinta
de impresoras Epson, portada desde la versión original en Python (`src/ez_reset`)
y la versión en Go (`resetgo/`).

## Características

- Interfaz gráfica Qt6 (modelos, búsqueda/filtro, detalle de estado, medidores de tinta y contadores de residuo).
- Detección automática de impresoras USB conectadas (solo Windows).
- Campo de conexión manual (ruta del puerto, p. ej. `\\.\USB001`).
- Lectura de estado, lectura/escritura de EEPROM, lectura y reset de contadores de residuo, limpieza, apagado y reinicio.
- Base de datos de modelos `devices.xml` (misma que la versión Python/Go).

## Requisitos

- C++17 o superior
- CMake 3.16+
- Qt6 (`qt6-base-dev`, `qt6-base-dev-tools`)
- En Windows: el transporte USB usa `setupapi` (se enlaza automáticamente).

## Compilar (Linux)

```bash
cd resetcplusplus
cmake -S . -B build -DCMAKE_PREFIX_PATH=/usr/lib/x86_64-linux-gnu/cmake
cmake --build build -j4
./build/bustamante_print_tools
```

> Nota: en Linux el transporte USBPRINT es solo un stub que lanza una excepción,
> porque la API de dispositivo USB de Windows no está disponible. La detección de
> impresoras y la conexión real solo funcionan en Windows. La GUI, la base de datos
> de modelos y la búsqueda funcionan en cualquier plataforma.

## Compilar (Windows)

Con Qt6 instalado y `setupapi` disponible (parte del SDK de Windows):

```powershell
cmake -S . -B build
cmake --build build --config Release
```

El ejecutable `bustamante_print_tools.exe` se enlaza con `setupapi.lib` y usa
`CreateFileW` / `DeviceIoControl` para hablar con la impresora vía el driver USBPRINT.

## Estructura

```
resetcplusplus/
├── CMakeLists.txt
├── devices.xml
├── include/
│   ├── utils.h        # parseo de estado/identificador
│   ├── status.h       # enums y Status
│   ├── devices.h      # carga de devices.xml
│   ├── transport.h    # interfaz de transporte + USBPRINT (Win/Linux)
│   ├── d4.h           # protocolo D4 (IEEE 1284.4)
│   ├── printer.h      # ControlBackend + Printer
│   └── mainwindow.h   # GUI Qt6
└── src/
    ├── main.cpp
    ├── utils.cpp
    ├── status.cpp
    ├── devices.cpp
    ├── transport.cpp
    ├── d4.cpp
    ├── printer.cpp
    └── mainwindow.cpp
```

## Notas

- El protocolo D4 es una reimplementación del transporte IEEE 1284.4 usado por las
  utilidades de mantenimiento de Epson.
- El reset de contadores de residuo (`Reset all waste counters`) escribe en la
  EEPROM de la impresora; úsalo bajo tu propia responsabilidad.
