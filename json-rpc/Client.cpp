//
// Created by chenlan on 2020/12/16.
//

#include "Client.h"

int JRpc::Client::increaseId = 0;

JRpc::Client::Client(const std::string &addr, const int port) {
//  memset(buffer, 0, sizeof(buffer));
  increaseId++;
  id = increaseId;

  memset(&serverAddr, 0, sizeof(serverAddr));
  serverAddr.sin_family = AF_INET;
  serverAddr.sin_addr.s_addr = inet_addr(addr.c_str());
  serverAddr.sin_port = htons(port);
  this->clientSocketFd = -1;
  this->version = "1.0";
}
bool JRpc::Client::call(const std::string &method, Json &params, Json &reply) {
  Json request{
      {"method", method},
      {"params", params},
      {"jsonrpc", version},
      {"id", id}
  };
  auto reqStr = request.dump();

  if (send(clientSocketFd, reqStr.c_str(), reqStr.size(), 0) < 0) {
    std::cerr << "Call remote service failed" << std::endl;
    return false;
  }

  // handle the reply
  memset(buffer, 0, sizeof(buffer));
  if (recv(clientSocketFd, buffer, BUFSIZ, 0) < 0) {
    std::cerr << "Call remote service failed" << std::endl;
    return false;
  }

  reply = Json::parse(buffer);

  return true;
}
bool JRpc::Client::tryConnect() {
  clientSocketFd = socket(AF_INET, SOCK_STREAM, 0);
  if (clientSocketFd < 0) {
    std::cerr << "Socket fail to create" << std::endl;
    return false;
  }
  if (connect(clientSocketFd, (sockaddr *) &serverAddr, sizeof(sockaddr)) < 0) {
    std::cerr << "Failed to connect to server" << std::endl;
    return false;
  }
  return true;

}
void JRpc::Client::disconnect() {
  Json p,r;
  call("_disconnect",p,r);
  close(clientSocketFd);
}
