#ifndef EZRESET_D4_H
#define EZRESET_D4_H

#include <string>
#include <vector>
#include <map>
#include <stdexcept>

#include "transport.h"
#include "printer.h"

namespace ezreset {

struct D4Packet {
    int psid;
    int ssid;
    int credit;
    int control;
    std::vector<unsigned char> payload;
};

enum class D4Command {
    Init = 0, OpenChannel = 1, CloseChannel = 2, Credit = 3,
    CreditRequest = 4, Exit = 8, GetSocketID = 9
};

class D4Error : public std::runtime_error {
public:
    explicit D4Error(const std::string& msg) : std::runtime_error(msg) {}
};

class D4Channel {
public:
    D4Channel(class D4* d4, int ssid) : d4_(d4), ssid_(ssid) {}

    void open();
    void close();
    void write(const std::vector<unsigned char>& data);
    D4Packet read();
    void ensureCredit();

    int ssid() const { return ssid_; }
    int psid() const { return psid_; }
    int mtu() const { return mtu_; }
    int txCredit() const { return txCredit_; }
    void addTxCredit(int n) { txCredit_ += n; }
    int rxCredit() const { return rxCredit_; }
    void addRxCredit(int n) { rxCredit_ += n; }
    void setPsid(int p) { psid_ = p; }
    void setMtu(int m) { mtu_ = m; }
    void setTxCredit(int c) { txCredit_ = c; }
    void pushQueue(const D4Packet& p) { rxQueue_.push_back(p); }
    bool queueEmpty() const { return rxQueue_.empty(); }
    D4Packet popQueue() {
        D4Packet p = rxQueue_.front();
        rxQueue_.erase(rxQueue_.begin());
        return p;
    }
    int rxMax() const { return rxMax_; }

private:
    class D4* d4_;
    int ssid_;
    int psid_ = -1;
    int mtu_ = 0;
    int txCredit_ = 0;
    int rxCredit_ = 0;
    int rxMax_ = 0x0001;
    std::vector<D4Packet> rxQueue_;
};

class D4 {
public:
    explicit D4(Transport* transport);
    ~D4();

    void writePacket(D4Channel* channel, const D4Packet& packet);
    D4Packet readPacket(D4Channel* channel);
    std::vector<unsigned char> command(D4Command cmd, const std::vector<unsigned char>& payload = {});
    void init();
    void exit();
    int getSocketID(const std::string& name);
    void openChannel(D4Channel* channel);
    void closeChannel(D4Channel* channel);
    void credit(D4Channel* channel, int amount);
    int creditRequest(D4Channel* channel, int amount = 0xFFFF);
    D4Channel* channel(const std::string& name);

    Transport* transport() { return transport_; }

private:
    void readNextPacket();
    int getFreePSID();

    Transport* transport_;
    std::map<int, D4Channel*> channels_;
};

// D4-based control backend implementing the high-level control interface.
class D4ControlBackend : public ControlBackend {
public:
    explicit D4ControlBackend(Transport* transport) : transport_(transport) {}
    ~D4ControlBackend();

    void open();
    void close();
    std::vector<unsigned char> send(const std::vector<unsigned char>& command);
    std::string identify();

private:
    Transport* transport_;
    D4* d4_ = nullptr;
    D4Channel* channel_ = nullptr;
};

} // namespace ezreset

#endif // EZRESET_D4_H
