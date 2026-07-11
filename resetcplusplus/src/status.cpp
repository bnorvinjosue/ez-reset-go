#include "status.h"
#include "utils.h"

#include <stdexcept>

namespace ezreset {

std::string printerStateToString(PrinterState s) {
    switch (s) {
        case PrinterState::SelfPrinting: return "SELF_PRINTING";
        case PrinterState::Busy: return "BUSY";
        case PrinterState::Waiting: return "WAITING";
        case PrinterState::Idle: return "IDLE";
        case PrinterState::Pause: return "PAUSE";
        case PrinterState::InkDrying: return "INKDRYING";
        case PrinterState::Cleaning: return "CLEANING";
        case PrinterState::FactoryShipment: return "FACTORY_SHIPMENT";
        case PrinterState::MotorDriveOff: return "MOTOR_DRIVE_OFF";
        case PrinterState::Shutdown: return "SHUTDOWN";
        case PrinterState::WaitPaperInit: return "WAITPAPERINIT";
        case PrinterState::InitPaper: return "INIT_PAPER";
        default: return "ERROR";
    }
}

std::string printerErrorToString(PrinterError e) {
    switch (e) {
        case PrinterError::None: return "NONE";
        case PrinterError::Fatal: return "FATAL";
        case PrinterError::Interface: return "INTERFACE";
        case PrinterError::PaperJam: return "PAPERJAM";
        case PrinterError::InkOut: return "INKOUT";
        case PrinterError::PaperOut: return "PAPEROUT";
        case PrinterError::PaperSize: return "PAPERSIZE";
        case PrinterError::PaperPath: return "PAPERPATH";
        case PrinterError::ServiceReq: return "SERVICEREQ";
        case PrinterError::DoubleFeed: return "DOUBLEFEED";
        case PrinterError::InkCoverOpen: return "INKCOVEROPEN";
        case PrinterError::NoMaintBox: return "NOMAINTENANCEBOX";
        case PrinterError::CoverOpen: return "COVEROPEN";
        case PrinterError::NoTray: return "NOTRAY";
        case PrinterError::CardLoading: return "CARDLOADING";
        case PrinterError::CDDVDConfig: return "CDDVDCONFIG";
        case PrinterError::CartridgeOverfl: return "CARTRIDGEOVERFLOW";
        case PrinterError::BatteryVoltage: return "BATTERYVOLTAGE";
        case PrinterError::BatteryTemp: return "BATTERYTEMPERATURE";
        case PrinterError::BatteryEmpty: return "BATTERYEMPTY";
        case PrinterError::Shutoff: return "SHUTOFF";
        case PrinterError::NotInitialFill: return "NOT_INITIALFILL";
        case PrinterError::PrintPackEnd: return "PRINTPACKEND";
        case PrinterError::MaintBoxOpen: return "MAINTENANCEBOXCOVEROPEN";
        case PrinterError::ScannerOpen: return "SCANNEROPEN";
        case PrinterError::CDRGuideOpen: return "CDRGUIDEOPEN";
        case PrinterError::CDRExist: return "CDREXIST";
        case PrinterError::CDRExistMainte: return "CDREXIST_MAINTE";
        case PrinterError::TrayClose: return "TRAYCLOSE";
        default: return "FATAL";
    }
}

std::string paperPathToString(PaperPath p) {
    switch (p) {
        case PaperPath::Roll: return "ROLL";
        case PaperPath::Fanfold: return "FANFOLD";
        case PaperPath::RollBack: return "ROLL_BACK";
        default: return "UNKNOWN";
    }
}

std::string consumableStatusToString(ConsumableStatus c) {
    switch (c) {
        case ConsumableStatus::Okay: return "OKAY";
        case ConsumableStatus::Empty: return "EMPTY";
        case ConsumableStatus::Missing: return "MISSING";
        case ConsumableStatus::Fail: return "FAIL";
        default: return "UNKNOWN";
    }
}

std::string inkColorToString(InkColor c) {
    switch (c) {
        case InkColor::Black: return "BLACK";
        case InkColor::Cyan: return "CYAN";
        case InkColor::Magenta: return "MAGENTA";
        case InkColor::Yellow: return "YELLOW";
        case InkColor::LightCyan: return "LIGHT_CYAN";
        case InkColor::LightMagenta: return "LIGHT_MAGENTA";
        case InkColor::DarkYellow: return "DARK_YELLOW";
        case InkColor::Gray: return "GRAY";
        case InkColor::LightBlack: return "LIGHT_BLACK";
        case InkColor::Red: return "RED";
        case InkColor::Blue: return "BLUE";
        case InkColor::GlossOptimizer: return "GLOSS_OPTIMIZER";
        case InkColor::LightGray: return "LIGHT_GRAY";
        case InkColor::Orange: return "ORANGE";
        default: return "UNKNOWN";
    }
}

ConsumableLevel consumableLevelFromInt(int level) {
    if (level == 110) return {-1, ConsumableStatus::Missing};
    if (level == 105) return {-1, ConsumableStatus::Unknown};
    if (level < 0 || level > 100) return {-1, ConsumableStatus::Fail};
    if (level == 0) return {0, ConsumableStatus::Empty};
    return {level, ConsumableStatus::Okay};
}

InkLevel inkLevelFromBytes(const std::vector<unsigned char>& entry) {
    InkLevel lvl;
    lvl.color = static_cast<InkColor>(entry[1]);
    lvl.consumable = consumableLevelFromInt(entry[2]);
    return lvl;
}

Status statusFromBytes(const std::vector<unsigned char>& data) {
    Status st;
    auto entries = parseStatusStruct(data);
    for (const auto& entry : entries) {
        switch (entry.header) {
            case 0x01:
                st.state = static_cast<PrinterState>(entry.payload[0]);
                break;
            case 0x02:
                st.error = static_cast<PrinterError>(entry.payload[0]);
                break;
            case 0x06:
                st.source = static_cast<PaperPath>(3 - entry.payload[0]);
                break;
            case 0x0D:
                st.maintenanceBox = consumableLevelFromInt(entry.payload[0]);
                break;
            case 0x0F: {
                int entrySize = entry.payload[0];
                if (entrySize <= 0) throw std::runtime_error("invalid ink entry size");
                for (size_t i = 1; i + entrySize <= entry.payload.size(); i += entrySize) {
                    std::vector<unsigned char> sub(entry.payload.begin() + i,
                                                   entry.payload.begin() + i + entrySize);
                    st.levels.push_back(inkLevelFromBytes(sub));
                }
                break;
            }
            case 0x40:
                st.serial.assign(entry.payload.begin(), entry.payload.end());
                break;
            default:
                st.other[entry.header] = entry.payload;
                break;
        }
    }
    return st;
}

} // namespace ezreset
