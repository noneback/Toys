//
// Created by chenlan on 2020/12/14.
//

#ifndef JSON_RPC__SERVER_H_
#define JSON_RPC__SERVER_H_
#include <string>
#include "json.h"
#include <iostream>
#include <unordered_map>
#include <functional>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <cstring>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

using MethodName = std::string;
using Json = nlohmann::json;

namespace JRpc {

enum ErrorType {
  JRPC_PARSE_ERROR = -32700,
  JRPC_INVALID_REQUEST = -32600,
  JRPC_METHOD_NOT_FOUND = -32601,
  JRPC_INVALID_PARAMS = -32603,
  JRPC_INTERNAL_ERROR = -32693,
  NO_ERROR = 0,
  AS_INFO
};

struct JRpcContext {
  nlohmann::json data;
  std::string errMsg;
  ErrorType errorCode;
};

struct JRpcConnection {
  int socketFd;
  sockaddr_in addr;
  char buffer[BUFSIZ];

};



enum ServerErrorType {
  SERVICE_HAS_REGISTERED,
  SOCKET_NOT_CREATE,
  SOCKET_BIND_ERROR,
  SOCKET_LISTEN_ERROR,
  SERVER_NO_ERROR,
  PARSE_REQUEST_ERROR
};

class Server {
 public:
  ~Server();
  Server(const int port);
  ServerErrorType regService(const MethodName &methodName, std::function<Json(const Json &)> func);
  void Close();
  JRpc::ServerErrorType serve();
  int servicesCount();

  ServerErrorType serveConn(JRpcConnection conn);

 private:
  ServerErrorType init();
  ServerErrorType sendError(int socketFd, ErrorType code, const std::string msg);
  JRpc::ServerErrorType sendResponse(int socketFd, Json result, int id);
  JRpcContext parseRequest(JRpc::JRpcConnection &conn);

  std::unordered_map<MethodName, std::function<Json(const Json &)>> servicesMap;
  int serverSocketFd;
  sockaddr_in serverAddr;
  std::string version;
  int port;
  bool active = false;
};
}

#endif //JSON_RPC__SERVER_H_
