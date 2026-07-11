#include "mainwindow.h"

#include <QListWidget>
#include <QLineEdit>
#include <QPushButton>
#include <QLabel>
#include <QProgressBar>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QGroupBox>
#include <QFrame>
#include <QMessageBox>
#include <QScrollArea>

#include "transport.h"
#include "d4.h"
#include "printer.h"
#include "utils.h"
#include "status.h"

namespace {

QString inkColorCss(const QString& name) {
    static const QMap<QString, QString> map = {
        {"BLACK", "#1f2937"}, {"CYAN", "#06b6d4"}, {"MAGENTA", "#db2777"},
        {"YELLOW", "#eab308"}, {"LIGHT_CYAN", "#67e8f9"}, {"LIGHT_MAGENTA", "#f9a8d4"},
        {"DARK_YELLOW", "#ca8a04"}, {"GRAY", "#9ca3af"}, {"LIGHT_BLACK", "#4b5563"},
        {"RED", "#ef4444"}, {"BLUE", "#3b82f6"}, {"GLOSS_OPTIMIZER", "#a78bfa"},
        {"LIGHT_GRAY", "#cbd5e1"}, {"ORANGE", "#f97316"}};
    return map.value(name, "#38bdf8");
}

} // namespace

MainWindow::MainWindow(QWidget* parent) : QMainWindow(parent) {
    setWindowTitle("Bustamante Print Tools");
    resize(960, 680);
    setStyleSheet(R"(
        QMainWindow { background: #0f1827; color: #e2e8f0; }
        QListWidget, QLineEdit { background: #1d2b48; color: #e2e8f0;
            border: 1px solid #243049; border-radius: 10px; padding: 6px; }
        QLineEdit:focus { border: 1px solid #38bdf8; }
        QPushButton { background: #38bdf8; color: #06121f; border: none;
            border-radius: 10px; padding: 9px 14px; font-weight: 600; }
        QPushButton:hover { background: #7dd3fc; }
        QPushButton:disabled { background: #334155; color: #94a3b8; }
        QGroupBox { border: 1px solid #243049; border-radius: 14px;
            margin-top: 10px; padding: 14px; }
        QGroupBox::title { subcontrol-origin: margin; left: 14px; color: #94a3b8;
            padding: 0 4px; }
        QLabel { color: #e2e8f0; }
    )");

    auto* central = new QWidget(this);
    setCentralWidget(central);
    auto* layout = new QHBoxLayout(central);
    layout->setSpacing(18);
    layout->setContentsMargins(18, 18, 18, 18);

    // ---- Left panel ----
    auto* left = new QVBoxLayout();
    left->setSpacing(10);

    auto* printerHead = new QHBoxLayout();
    printerHead->addWidget(new QLabel("<b>Printers</b>"));
    auto* refreshBtn = new QPushButton("⟳");
    refreshBtn->setMaximumWidth(40);
    printerHead->addWidget(refreshBtn);
    left->addLayout(printerHead);

    printerList_ = new QListWidget();
    printerList_->setMaximumHeight(180);
    left->addWidget(printerList_);

    auto* manualHead = new QHBoxLayout();
    manualHead->addWidget(new QLabel("<b>Connect manually</b>"));
    left->addLayout(manualHead);
    manualPath_ = new QLineEdit();
    manualPath_->setPlaceholderText("e.g. \\\\.\\USBPRINT\\Epson...");
    left->addWidget(manualPath_);
    connectBtn_ = new QPushButton("Connect");
    left->addWidget(connectBtn_);

    auto* modelHead = new QHBoxLayout();
    modelHead->addWidget(new QLabel("<b>Supported models</b>"));
    modelCount_ = new QLabel("0");
    modelHead->addWidget(modelCount_);
    left->addLayout(modelHead);
    modelSearch_ = new QLineEdit();
    modelSearch_->setPlaceholderText("Search models…");
    left->addWidget(modelSearch_);
    modelList_ = new QListWidget();
    modelList_->setMaximumHeight(220);
    left->addWidget(modelList_);

    auto* leftWidget = new QWidget();
    leftWidget->setLayout(left);
    leftWidget->setMaximumWidth(340);
    layout->addWidget(leftWidget);

    // ---- Right panel ----
    detailWidget_ = new QWidget();
    auto* right = new QVBoxLayout(detailWidget_);
    right->setSpacing(14);

    auto* detailHead = new QHBoxLayout();
    printerName_ = new QLabel("—");
    printerName_->setStyleSheet("font-size: 22px; font-weight: 700;");
    printerMeta_ = new QLabel("—");
    printerMeta_->setStyleSheet("color: #94a3b8;");
    stateBadge_ = new QLabel("—");
    stateBadge_->setStyleSheet("padding: 4px 12px; border-radius: 999px; border: 1px solid #243049;");
    detailHead->addWidget(printerName_);
    detailHead->addWidget(printerMeta_, 1);
    detailHead->addWidget(stateBadge_);
    right->addLayout(detailHead);

    auto* cards = new QHBoxLayout();
    cards->setSpacing(16);

    auto* inkBox = new QGroupBox("Ink levels");
    inkLayout_ = new QVBoxLayout(inkBox);
    cards->addWidget(inkBox);

    auto* wasteBox = new QGroupBox("Waste ink counters");
    wasteLayout_ = new QVBoxLayout(wasteBox);
    resetBtn_ = new QPushButton("Reset all waste counters");
    resetBtn_->setStyleSheet("background: #f43f5e; color: white;");
    wasteLayout_->addWidget(resetBtn_);
    resetMsg_ = new QLabel("");
    resetMsg_->setStyleSheet("color: #34d399;");
    wasteLayout_->addWidget(resetMsg_);
    cards->addWidget(wasteBox);

    right->addLayout(cards);
    right->addStretch(1);

    layout->addWidget(detailWidget_, 1);

    connect(refreshBtn, &QPushButton::clicked, this, &MainWindow::refreshPrinters);
    connect(printerList_, &QListWidget::itemClicked, [this](QListWidgetItem* item) {
        onPrinterSelected(item->data(Qt::UserRole).toString(), item->text());
    });
    connect(connectBtn_, &QPushButton::clicked, this, &MainWindow::connectManual);
    connect(resetBtn_, &QPushButton::clicked, this, &MainWindow::resetWaste);
    connect(modelSearch_, &QLineEdit::textChanged, [this](const QString& text) {
        onModelSearch(text);
    });

    loadModels();
    refreshPrinters();
}

void MainWindow::loadModels() {
    // Locate devices.xml next to the binary or in known locations.
    QStringList candidates = {"devices.xml", "resetcplusplus/devices.xml",
                              "src/ez_reset/devices.xml", "resetgo/internal/devices/devices.xml"};
    for (const auto& c : candidates) {
        try {
            devices_ = ezreset::loadDevices(c.toStdString());
            break;
        } catch (...) {}
    }
    allModels_.clear();
    for (const auto& kv : devices_) {
        allModels_.append(QString::fromStdString(kv.first));
    }
    allModels_.sort();
    renderModels("");
}

void MainWindow::renderModels(const QString& filter) {
    modelList_->clear();
    int shown = 0;
    QString q = filter.toLower();
    for (const auto& m : allModels_) {
        if (!q.isEmpty() && !m.toLower().contains(q)) continue;
        modelList_->addItem(m);
        shown++;
    }
    modelCount_->setText(QString::number(shown));
}

void MainWindow::onModelSearch(const QString& text) {
    renderModels(text);
}

void MainWindow::refreshPrinters() {
    printerList_->clear();
    try {
        auto paths = ezreset::enumeratePrinters();
        for (const auto& p : paths) {
            auto* t = new ezreset::USBPRINTTransport(p);
            QString label = QString::fromStdString(p);
            try {
                t->open();
                std::string id = t->identify();
                std::string mdl = ezreset::parseField(id, "MDL");
                std::string sn = ezreset::parseField(id, "SN");
                if (!mdl.empty()) label = QString::fromStdString(mdl);
                if (!sn.empty()) label += "  (S/N: " + QString::fromStdString(sn) + ")";
                t->close();
            } catch (...) {}
            delete t;
            auto* item = new QListWidgetItem(label);
            item->setData(Qt::UserRole, QString::fromStdString(p));
            printerList_->addItem(item);
        }
        if (paths.empty()) {
            printerList_->addItem("No USB printers found.");
        }
    } catch (const std::exception& e) {
        printerList_->addItem(QString("Enumeration failed: ") + e.what());
    }
}

void MainWindow::onPrinterSelected(const QString& path, const QString& label) {
    if (path.isEmpty()) return;
    showStatus(path, label);
}

void MainWindow::connectManual() {
    QString path = manualPath_->text().trimmed();
    if (path.isEmpty()) {
        QMessageBox::warning(this, "Connect", "Enter a printer path to connect.");
        return;
    }
    showStatus(path, path);
}

void MainWindow::showStatus(const QString& path, const QString& label) {
    printerName_->setText(label);
    printerMeta_->setText(path);

    auto* t = new ezreset::USBPRINTTransport(path.toStdString());
    try {
        t->open();
    } catch (const std::exception& e) {
        QMessageBox::critical(this, "Error", QString("Failed to open printer: ") + e.what());
        delete t;
        return;
    }

    ezreset::D4ControlBackend backend(t);
    try {
        backend.open();
    } catch (const std::exception& e) {
        QMessageBox::critical(this, "Error", QString("Failed to open control channel: ") + e.what());
        delete t;
        return;
    }

    try {
        std::string id = backend.identify();
        std::string model = ezreset::parseField(id, "MDL");
        auto device = ezreset::deviceByModel(devices_, model);

        ezreset::Printer printer(&backend, device);
        auto st = printer.getStatus();

        stateBadge_->setText(QString::fromStdString(ezreset::printerStateToString(st.state)));
        stateBadge_->setStyleSheet("padding: 4px 12px; border-radius: 999px; border: 1px solid #38bdf8; color: #38bdf8;");

        // Ink levels
        QLayoutItem* child;
        while ((child = inkLayout_->takeAt(0)) != nullptr) { delete child->widget(); delete child; }
        for (const auto& lvl : st.levels) {
            auto* row = new QHBoxLayout();
            auto* name = new QLabel(QString::fromStdString(ezreset::inkColorToString(lvl.color)));
            auto* bar = new QProgressBar();
            bar->setRange(0, 100);
            bar->setValue(lvl.consumable.level);
            bar->setFormat("%v%");
            row->addWidget(name, 1);
            row->addWidget(bar, 2);
            auto* w = new QWidget();
            w->setLayout(row);
            inkLayout_->addWidget(w);
        }
        if (st.levels.empty()) inkLayout_->addWidget(new QLabel("No ink data."));

        // Waste counters
        while ((child = wasteLayout_->takeAt(0)) != nullptr) {
            if (child->widget() && (child->widget() == resetBtn_ || child->widget() == resetMsg_)) continue;
            delete child->widget(); delete child;
        }
        auto wastes = printer.getWaste();
        for (size_t i = 0; i < wastes.size(); i++) {
            int value = wastes[i].first, max = wastes[i].second;
            double ratio = max > 0 ? static_cast<double>(value) / max : 0.0;
            auto* row = new QHBoxLayout();
            auto* name = new QLabel(QString("Counter %1").arg(i));
            auto* bar = new QProgressBar();
            bar->setRange(0, 100);
            bar->setValue(static_cast<int>(ratio * 100));
            bar->setFormat(QString("%1 / %2 (%3%)").arg(value).arg(max).arg(static_cast<int>(ratio * 100)));
            if (ratio > 0.8) bar->setStyleSheet("QProgressBar::chunk { background: #f43f5e; }");
            row->addWidget(name, 1);
            row->addWidget(bar, 2);
            auto* w = new QWidget();
            w->setLayout(row);
            wasteLayout_->insertWidget(wasteLayout_->count() - 2, w);
        }
        resetMsg_->clear();
        resetBtn_->setProperty("path", path);
    } catch (const std::exception& e) {
        QMessageBox::critical(this, "Error", QString("Operation failed: ") + e.what());
    }

    backend.close();
    delete t;
}

void MainWindow::resetWaste() {
    QString path = resetBtn_->property("path").toString();
    if (path.isEmpty()) return;

    auto* t = new ezreset::USBPRINTTransport(path.toStdString());
    try { t->open(); } catch (const std::exception& e) {
        QMessageBox::critical(this, "Error", QString("Failed to open printer: ") + e.what());
        delete t; return;
    }
    ezreset::D4ControlBackend backend(t);
    try { backend.open(); } catch (const std::exception& e) {
        QMessageBox::critical(this, "Error", QString("Failed to open control channel: ") + e.what());
        delete t; return;
    }
    try {
        std::string id = backend.identify();
        std::string model = ezreset::parseField(id, "MDL");
        auto device = ezreset::deviceByModel(devices_, model);
        ezreset::Printer printer(&backend, device);
        printer.resetWaste();
        resetMsg_->setText("Waste ink counters have been reset. Restart the printer.");
        QMessageBox::information(this, "Reset", "Waste ink counters have been reset. Restart the printer.");
        showStatus(path, path);
    } catch (const std::exception& e) {
        QMessageBox::critical(this, "Error", QString("Reset failed: ") + e.what());
    }
    backend.close();
    delete t;
}
