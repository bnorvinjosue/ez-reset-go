#ifndef EZRESET_STATUS_H
#define EZRESET_STATUS_H

#include <string>
#include <vector>
#include <map>

namespace ezreset {

enum class PrinterState {
    Error = 0x00, SelfPrinting = 0x01, Busy = 0x02, Waiting = 0x03,
    Idle = 0x04, Pause = 0x05, InkDrying = 0x06, Cleaning = 0x07,
    FactoryShipment = 0x08, MotorDriveOff = 0x09, Shutdown = 0x0A,
    WaitPaperInit = 0x0B, InitPaper = 0x0C
};

enum class PrinterError {
    None = -1, Fatal = 0x00, Interface = 0x01, PaperJam = 0x04,
    InkOut = 0x05, PaperOut = 0x06, PaperSize = 0x0A, PaperPath = 0x0C,
    ServiceReq = 0x10, DoubleFeed = 0x12, InkCoverOpen = 0x1A,
    NoMaintBox = 0x22, CoverOpen = 0x25, NoTray = 0x29, CardLoading = 0x2A,
    CDDVDConfig = 0x2B, CartridgeOverfl = 0x2C, BatteryVoltage = 0x2F,
    BatteryTemp = 0x30, BatteryEmpty = 0x31, Shutoff = 0x32,
    NotInitialFill = 0x33, PrintPackEnd = 0x34, MaintBoxOpen = 0x36,
    ScannerOpen = 0x37, CDRGuideOpen = 0x38, CDRExist = 0x44,
    CDRExistMainte = 0x45, TrayClose = 0x46
};

enum class PaperPath {
    Unknown = -1, Roll = 0x00, Fanfold = 0x01, RollBack = 0x02
};

enum class ConsumableStatus {
    Okay = 0, Empty = 1, Missing = 2, Fail = 3, Unknown = 4
};

enum class InkColor {
    Black = 0, Cyan = 1, Magenta = 2, Yellow = 3, LightCyan = 4,
    LightMagenta = 5, DarkYellow = 6, Gray = 7, LightBlack = 8, Red = 9,
    Blue = 10, GlossOptimizer = 11, LightGray = 12, Orange = 13, Unknown = -1
};

struct ConsumableLevel {
    int level;
    ConsumableStatus status;
};

struct InkLevel {
    ConsumableLevel consumable;
    InkColor color;
};

struct Status {
    PrinterState state = PrinterState::Idle;
    PrinterError error = PrinterError::None;
    PaperPath source = PaperPath::Unknown;
    std::vector<InkLevel> levels;
    ConsumableLevel maintenanceBox{-1, ConsumableStatus::Unknown};
    std::string serial;
    std::map<int, std::vector<unsigned char>> other;
};

std::string printerStateToString(PrinterState s);
std::string printerErrorToString(PrinterError e);
std::string paperPathToString(PaperPath p);
std::string consumableStatusToString(ConsumableStatus c);
std::string inkColorToString(InkColor c);

ConsumableLevel consumableLevelFromInt(int level);
InkLevel inkLevelFromBytes(const std::vector<unsigned char>& entry);

// Parse a Status from the binary payload following the "@BDC ST2" header.
Status statusFromBytes(const std::vector<unsigned char>& data);

} // namespace ezreset

#endif // EZRESET_STATUS_H
