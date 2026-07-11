#include "devices.h"

#include <QFile>
#include <QXmlStreamReader>
#include <QString>
#include <stdexcept>

namespace ezreset {

namespace {

std::vector<unsigned char> parseByteList(const QString& text) {
    std::vector<unsigned char> out;
    for (const QString& f : text.split(' ', Qt::SkipEmptyParts)) {
        bool ok = false;
        int v = f.toInt(&ok, 0);
        if (ok) out.push_back(static_cast<unsigned char>(v));
    }
    return out;
}

std::vector<int> parseAddressList(const QString& text) {
    std::vector<int> out;
    for (const QString& f : text.split(' ', Qt::SkipEmptyParts)) {
        bool ok = false;
        int v = f.toInt(&ok, 0);
        if (ok) out.push_back(v);
    }
    return out;
}

} // namespace

std::map<std::string, Device> loadDevices(const std::string& path) {
    QFile file(QString::fromStdString(path));
    if (!file.open(QIODevice::ReadOnly)) {
        throw std::runtime_error("cannot open devices.xml: " + path);
    }
    QXmlStreamReader xml(&file);

    std::map<std::string, Device> byModel;
    std::map<std::string, Device> specs; // named <device> entries keyed by tag

    bool inRecords = false;
    bool inDevices = false;
    QString currentSpecName;

    while (!xml.atEnd()) {
        xml.readNext();
        if (xml.isStartElement()) {
            QString name = xml.name().toString();
            if (name == "records") {
                inRecords = true;
            } else if (name == "devices") {
                inDevices = true;
            } else if (inRecords && name == "printer") {
                QXmlStreamAttributes a = xml.attributes();
                QString model = a.value("model").toString().toUtf8().constData();
                QString specsAttr = a.value("specs").toString();
                Device dev;
                dev.reset = {};
                for (const QString& specName : specsAttr.split(',', Qt::SkipEmptyParts)) {
                    QString s = specName.trimmed();
                    if (s.isEmpty()) continue;
                    auto it = specs.find(s.toUtf8().constData());
                    if (it == specs.end()) continue;
                    const Device& spec = it->second;
                    if (!spec.model.empty()) dev.model = spec.model;
                    if (!spec.key.empty()) dev.key = spec.key;
                    for (const auto& c : spec.counters) dev.counters.push_back(c);
                    for (const auto& kv : spec.reset) dev.reset[kv.first] = kv.second;
                }
                if (!model.isEmpty() && model != "Device") {
                    byModel[model.toUtf8().constData()] = dev;
                }
            } else if (inDevices) {
                currentSpecName = name; // e.g. SC700
                specs[currentSpecName.toUtf8().constData()] = Device{};
            } else if (!currentSpecName.isEmpty()) {
                // Inside a named device spec.
                Device& spec = specs[currentSpecName.toUtf8().constData()];
                if (name == "factory") {
                    spec.model = parseByteList(xml.readElementText());
                } else if (name == "keyword") {
                    spec.key = parseByteList(xml.readElementText());
                } else if (name == "counter") {
                    Counter c;
                    QString txt = xml.readElementText().trimmed();
                    c.addresses = parseAddressList(txt);
                    spec.counters.push_back(c);
                } else if (name == "max") {
                    if (!spec.counters.empty()) {
                        spec.counters.back().max = xml.readElementText().trimmed().toInt();
                    }
                } else if (name == "reset") {
                    QStringList fields = xml.readElementText().split(' ', Qt::SkipEmptyParts);
                    for (int i = 0; i + 1 < fields.size(); i += 2) {
                        bool ok1 = false, ok2 = false;
                        int addr = fields[i].toInt(&ok1, 0);
                        int val = fields[i + 1].toInt(&ok2, 0);
                        if (ok1 && ok2) spec.reset[addr] = val;
                    }
                }
            }
        } else if (xml.isEndElement()) {
            QString ename = xml.name().toString();
            if (ename == "records") inRecords = false;
            else if (ename == "devices") inDevices = false;
            else if (inDevices && !currentSpecName.isEmpty()
                     && ename == currentSpecName) {
                currentSpecName.clear();
            }
        }
    }
    if (xml.hasError()) {
        throw std::runtime_error("XML parse error in devices.xml");
    }
    return byModel;
}

Device deviceByModel(const std::map<std::string, Device>& devices, const std::string& model) {
    auto it = devices.find(model);
    if (it == devices.end()) {
        throw std::runtime_error("unknown printer model: " + model);
    }
    return it->second;
}

} // namespace ezreset
