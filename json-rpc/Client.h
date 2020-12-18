//
// Created by chenlan on 2020/12/16.
//

#ifndef JSON_RPC__CLIENT_H_
#define JSON_RPC__CLIENT_H_

#include <string>
#include "json.h"
#include <iostream>
#include <unordered_map>
#include <functional>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>
using Json = nlohmann::json;
namespace JRpc {
class Client {
 public:
  Client(const std::string &addr, const int port);
  //~Client();
  void disconnect();
  bool tryConnect();
  bool call(const std::string &method, Json &params, Json &reply);
 private:
  static int increaseId;
  int clientSocketFd;
  int id;
  sockaddr_in serverAddr;
  std::string version;
  char buffer[BUFSIZ];
  bool init();

};

}

#endif //JSON_RPC__CLIENT_H_
