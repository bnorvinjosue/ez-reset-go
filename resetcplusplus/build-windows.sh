#!/usr/bin/env bash
# Cross-compile Bustamante Print Tools (C++/Qt6) for Windows from Linux.
#
# Prerequisites:
#   sudo apt-get install -y mingw-w64
#   pip install aqtinstall
#   python3 -m aqt install-qt windows desktop 6.7.0 win64_mingw   # downloads to ./6.7.0/mingw_64
#
# Usage:
#   ./build-windows.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$HERE"

QT_PREFIX="${QT_PREFIX:-$HERE/6.7.0/mingw_64}"
MOC="${MOC:-/usr/lib/qt6/libexec/moc}"
RCC="${RCC:-/usr/lib/qt6/libexec/rcc}"
UIC="${UIC:-/usr/lib/qt6/libexec/uic}"

rm -rf build-win
cmake -S . -B build-win \
  -DCMAKE_TOOLCHAIN_FILE=toolchain-mingw.cmake \
  -DCMAKE_PREFIX_PATH="$QT_PREFIX" \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_AUTOMOC_EXECUTABLE="$MOC" \
  -DCMAKE_AUTORCC_EXECUTABLE="$RCC" \
  -DCMAKE_AUTOUIC_EXECUTABLE="$UIC" \
  -DCMAKE_DISABLE_FIND_PACKAGE_WrapVulkanHeaders=ON

cmake --build build-win -j"$(nproc)"

# Assemble a redistributable folder with the required DLLs.
rm -rf dist-win
mkdir -p dist-win/platforms
cp build-win/bustamante_print_tools.exe dist-win/
cp build-win/devices.xml dist-win/
cp "$QT_PREFIX"/bin/Qt6Core.dll "$QT_PREFIX"/bin/Qt6Gui.dll "$QT_PREFIX"/bin/Qt6Widgets.dll dist-win/
cp "$QT_PREFIX"/plugins/platforms/qwindows.dll dist-win/platforms/
MINGW_LIB="/usr/lib/gcc/x86_64-w64-mingw32/$(ls /usr/lib/gcc/x86_64-w64-mingw32/)/"
cp "$MINGW_LIB"/libgcc_s_seh-1.dll "$MINGW_LIB"/libstdc++-6.dll dist-win/

echo "Windows build ready in dist-win/"
