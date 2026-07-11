#ifndef EZRESET_MAINWINDOW_H
#define EZRESET_MAINWINDOW_H

#include <QMainWindow>
#include <QMap>
#include <QStringList>
#include <map>
#include <string>

#include "devices.h"

QT_BEGIN_NAMESPACE
class QListWidget;
class QLineEdit;
class QPushButton;
class QLabel;
class QProgressBar;
class QVBoxLayout;
class QHBoxLayout;
class QWidget;
QT_END_NAMESPACE

class MainWindow : public QMainWindow {
    Q_OBJECT

public:
    explicit MainWindow(QWidget* parent = nullptr);

private slots:
    void refreshPrinters();
    void onPrinterSelected(const QString& path, const QString& label);
    void connectManual();
    void resetWaste();
    void onModelSearch(const QString& text);

private:
    void loadModels();
    void renderModels(const QString& filter);
    void showStatus(const QString& path, const QString& model);

    std::map<std::string, ezreset::Device> devices_;
    QStringList allModels_;

    QListWidget* printerList_;
    QLineEdit* manualPath_;
    QPushButton* connectBtn_;
    QLineEdit* modelSearch_;
    QListWidget* modelList_;
    QLabel* modelCount_;

    QWidget* detailWidget_;
    QLabel* printerName_;
    QLabel* printerMeta_;
    QLabel* stateBadge_;
    QVBoxLayout* inkLayout_;
    QVBoxLayout* wasteLayout_;
    QLabel* resetMsg_;
    QPushButton* resetBtn_;
};

#endif // EZRESET_MAINWINDOW_H
