// Package status contains enums and structs related to printer status
// conditions and states, ported from the Python ez-reset tool.
//
// The enum values have been extracted from the open-source epson-inkjet-escpr
// driver package, released under GPLv2.
package status

import (
	"fmt"

	"ezreset/internal/utils"
)

// PrinterState represents the high-level printer state.
type PrinterState int

const (
	PrinterStateError          PrinterState = 0x00
	PrinterStateSelfPrinting   PrinterState = 0x01
	PrinterStateBusy           PrinterState = 0x02
	PrinterStateWaiting        PrinterState = 0x03
	PrinterStateIdle           PrinterState = 0x04
	PrinterStatePause          PrinterState = 0x05
	PrinterStateInkDrying      PrinterState = 0x06
	PrinterStateCleaning       PrinterState = 0x07
	PrinterStateFactoryShipment PrinterState = 0x08
	PrinterStateMotorDriveOff  PrinterState = 0x09
	PrinterStateShutdown       PrinterState = 0x0A
	PrinterStateWaitPaperInit  PrinterState = 0x0B
	PrinterStateInitPaper      PrinterState = 0x0C
)

func (s PrinterState) String() string {
	switch s {
	case PrinterStateSelfPrinting:
		return "SELF_PRINTING"
	case PrinterStateBusy:
		return "BUSY"
	case PrinterStateWaiting:
		return "WAITING"
	case PrinterStateIdle:
		return "IDLE"
	case PrinterStatePause:
		return "PAUSE"
	case PrinterStateInkDrying:
		return "INKDRYING"
	case PrinterStateCleaning:
		return "CLEANING"
	case PrinterStateFactoryShipment:
		return "FACTORY_SHIPMENT"
	case PrinterStateMotorDriveOff:
		return "MOTOR_DRIVE_OFF"
	case PrinterStateShutdown:
		return "SHUTDOWN"
	case PrinterStateWaitPaperInit:
		return "WAITPAPERINIT"
	case PrinterStateInitPaper:
		return "INIT_PAPER"
	default:
		return "ERROR"
	}
}

// PrinterError represents a printer error condition.
type PrinterError int

const (
	PrinterErrorNone            PrinterError = -1
	PrinterErrorFatal           PrinterError = 0x00
	PrinterErrorInterface       PrinterError = 0x01
	PrinterErrorPaperJam        PrinterError = 0x04
	PrinterErrorInkOut          PrinterError = 0x05
	PrinterErrorPaperOut        PrinterError = 0x06
	PrinterErrorPaperSize       PrinterError = 0x0A
	PrinterErrorPaperPath       PrinterError = 0x0C
	PrinterErrorServiceReq      PrinterError = 0x10
	PrinterErrorDoubleFeed      PrinterError = 0x12
	PrinterErrorInkCoverOpen    PrinterError = 0x1A
	PrinterErrorNoMaintBox      PrinterError = 0x22
	PrinterErrorCoverOpen       PrinterError = 0x25
	PrinterErrorNoTray          PrinterError = 0x29
	PrinterErrorCardLoading     PrinterError = 0x2A
	PrinterErrorCDDVDConfig     PrinterError = 0x2B
	PrinterErrorCartridgeOverfl PrinterError = 0x2C
	PrinterErrorBatteryVoltage  PrinterError = 0x2F
	PrinterErrorBatteryTemp     PrinterError = 0x30
	PrinterErrorBatteryEmpty    PrinterError = 0x31
	PrinterErrorShutoff         PrinterError = 0x32
	PrinterErrorNotInitialFill  PrinterError = 0x33
	PrinterErrorPrintPackEnd    PrinterError = 0x34
	PrinterErrorMaintBoxOpen    PrinterError = 0x36
	PrinterErrorScannerOpen     PrinterError = 0x37
	PrinterErrorCDRGuideOpen    PrinterError = 0x38
	PrinterErrorCDRExist        PrinterError = 0x44
	PrinterErrorCDRExistMainte  PrinterError = 0x45
	PrinterErrorTrayClose       PrinterError = 0x46
)

func (e PrinterError) String() string {
	switch e {
	case PrinterErrorNone:
		return "NONE"
	case PrinterErrorFatal:
		return "FATAL"
	case PrinterErrorInterface:
		return "INTERFACE"
	case PrinterErrorPaperJam:
		return "PAPERJAM"
	case PrinterErrorInkOut:
		return "INKOUT"
	case PrinterErrorPaperOut:
		return "PAPEROUT"
	case PrinterErrorPaperSize:
		return "PAPERSIZE"
	case PrinterErrorPaperPath:
		return "PAPERPATH"
	case PrinterErrorServiceReq:
		return "SERVICEREQ"
	case PrinterErrorDoubleFeed:
		return "DOUBLEFEED"
	case PrinterErrorInkCoverOpen:
		return "INKCOVEROPEN"
	case PrinterErrorNoMaintBox:
		return "NOMAINTENANCEBOX"
	case PrinterErrorCoverOpen:
		return "COVEROPEN"
	case PrinterErrorNoTray:
		return "NOTRAY"
	case PrinterErrorCardLoading:
		return "CARDLOADING"
	case PrinterErrorCDDVDConfig:
		return "CDDVDCONFIG"
	case PrinterErrorCartridgeOverfl:
		return "CARTRIDGEOVERFLOW"
	case PrinterErrorBatteryVoltage:
		return "BATTERYVOLTAGE"
	case PrinterErrorBatteryTemp:
		return "BATTERYTEMPERATURE"
	case PrinterErrorBatteryEmpty:
		return "BATTERYEMPTY"
	case PrinterErrorShutoff:
		return "SHUTOFF"
	case PrinterErrorNotInitialFill:
		return "NOT_INITIALFILL"
	case PrinterErrorPrintPackEnd:
		return "PRINTPACKEND"
	case PrinterErrorMaintBoxOpen:
		return "MAINTENANCEBOXCOVEROPEN"
	case PrinterErrorScannerOpen:
		return "SCANNEROPEN"
	case PrinterErrorCDRGuideOpen:
		return "CDRGUIDEOPEN"
	case PrinterErrorCDRExist:
		return "CDREXIST"
	case PrinterErrorCDRExistMainte:
		return "CDREXIST_MAINTE"
	case PrinterErrorTrayClose:
		return "TRAYCLOSE"
	default:
		return "FATAL"
	}
}

// PaperPath represents the media source / paper path.
type PaperPath int

const (
	PaperPathUnknown PaperPath = -1
	PaperPathRoll    PaperPath = 0x00
	PaperPathFanfold PaperPath = 0x01
	PaperPathRollBack PaperPath = 0x02
)

func (p PaperPath) String() string {
	switch p {
	case PaperPathRoll:
		return "ROLL"
	case PaperPathFanfold:
		return "FANFOLD"
	case PaperPathRollBack:
		return "ROLL_BACK"
	default:
		return "UNKNOWN"
	}
}

// ConsumableStatus represents the status of a consumable (ink/maintenance box).
type ConsumableStatus int

const (
	ConsumableStatusOkay    ConsumableStatus = 0
	ConsumableStatusEmpty   ConsumableStatus = 1
	ConsumableStatusMissing ConsumableStatus = 2
	ConsumableStatusFail    ConsumableStatus = 3
	ConsumableStatusUnknown ConsumableStatus = 4
)

func (c ConsumableStatus) String() string {
	switch c {
	case ConsumableStatusOkay:
		return "OKAY"
	case ConsumableStatusEmpty:
		return "EMPTY"
	case ConsumableStatusMissing:
		return "MISSING"
	case ConsumableStatusFail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// ConsumableLevel holds a consumable level value and its status.
type ConsumableLevel struct {
	Level  int
	Status ConsumableStatus
}

// FromInt parses level and status information from a raw level value.
func ConsumableLevelFromInt(level int) ConsumableLevel {
	switch {
	case level == 110:
		return ConsumableLevel{Level: -1, Status: ConsumableStatusMissing}
	case level == 105:
		return ConsumableLevel{Level: -1, Status: ConsumableStatusUnknown}
	case level < 0 || level > 100:
		return ConsumableLevel{Level: -1, Status: ConsumableStatusFail}
	case level == 0:
		return ConsumableLevel{Level: level, Status: ConsumableStatusEmpty}
	default:
		return ConsumableLevel{Level: level, Status: ConsumableStatusOkay}
	}
}

// InkColor represents an ink cartridge color.
type InkColor int

const (
	InkColorBlack          InkColor = 0
	InkColorCyan           InkColor = 1
	InkColorMagenta        InkColor = 2
	InkColorYellow         InkColor = 3
	InkColorLightCyan      InkColor = 4
	InkColorLightMagenta   InkColor = 5
	InkColorDarkYellow     InkColor = 6
	InkColorGray           InkColor = 7
	InkColorLightBlack     InkColor = 8
	InkColorRed            InkColor = 9
	InkColorBlue           InkColor = 10
	InkColorGlossOptimizer InkColor = 11
	InkColorLightGray      InkColor = 12
	InkColorOrange         InkColor = 13
	InkColorUnknown        InkColor = -1
)

func (c InkColor) String() string {
	switch c {
	case InkColorBlack:
		return "BLACK"
	case InkColorCyan:
		return "CYAN"
	case InkColorMagenta:
		return "MAGENTA"
	case InkColorYellow:
		return "YELLOW"
	case InkColorLightCyan:
		return "LIGHT_CYAN"
	case InkColorLightMagenta:
		return "LIGHT_MAGENTA"
	case InkColorDarkYellow:
		return "DARK_YELLOW"
	case InkColorGray:
		return "GRAY"
	case InkColorLightBlack:
		return "LIGHT_BLACK"
	case InkColorRed:
		return "RED"
	case InkColorBlue:
		return "BLUE"
	case InkColorGlossOptimizer:
		return "GLOSS_OPTIMIZER"
	case InkColorLightGray:
		return "LIGHT_GRAY"
	case InkColorOrange:
		return "ORANGE"
	default:
		return "UNKNOWN"
	}
}

// InkLevel holds an ink level value, status and color.
type InkLevel struct {
	ConsumableLevel
	Color InkColor
}

// InkLevelFromBytes parses an ink entry from an ink level field.
func InkLevelFromBytes(entry []byte) InkLevel {
	color := InkColor(int(entry[1]))
	consumable := ConsumableLevelFromInt(int(entry[2]))

	return InkLevel{
		ConsumableLevel: consumable,
		Color:           color,
	}
}

// Status is the parsed printer status.
type Status struct {
	State         PrinterState
	Error         PrinterError
	Source        PaperPath
	Levels        []InkLevel
	MaintenanceBox ConsumableLevel
	Serial        string
	Other         map[int][]byte
}

// FromBytes parses a Status from the binary payload following the "@BDC ST2" header.
func StatusFromBytes(data []byte) (*Status, error) {
	st := &Status{
		State:         PrinterStateIdle,
		Error:         PrinterErrorNone,
		Source:        PaperPathUnknown,
		MaintenanceBox: ConsumableLevel{Level: -1, Status: ConsumableStatusUnknown},
		Other:         make(map[int][]byte),
	}

	entries, err := utils.ParseStatusStruct(data)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		switch entry.Header {
		case 0x01: // Status entry
			st.State = PrinterState(int(entry.Payload[0]))
		case 0x02: // Error entry
			st.Error = PrinterError(int(entry.Payload[0]))
		case 0x06: // Media source entry
			st.Source = PaperPath(3 - int(entry.Payload[0]))
		case 0x0D: // Maintenance box level entry
			st.MaintenanceBox = ConsumableLevelFromInt(int(entry.Payload[0]))
		case 0x0F: // Ink entry
			entrySize := int(entry.Payload[0])
			if entrySize <= 0 {
				return nil, fmt.Errorf("invalid ink entry size %d", entrySize)
			}

			for i := 1; i+entrySize <= len(entry.Payload); i += entrySize {
				st.Levels = append(st.Levels, InkLevelFromBytes(entry.Payload[i:i+entrySize]))
			}
		case 0x40: // Serial number entry
			st.Serial = string(entry.Payload)
		default:
			st.Other[int(entry.Header)] = entry.Payload
		}
	}

	return st, nil
}
