//
// Created by chenlan on 2020/12/14.
//

#include "Server.h"
#include <thread>

/**
 * init the server socket
 */




JRpc::ServerErrorType JRpc::Server::init() {
  // init net family
  memset(&serverAddr, 0, sizeof(serverAddr));
  serverAddr.sin_family = AF_INET;
  serverAddr.sin_addr.s_addr = INADDR_ANY;
  serverAddr.sin_port = htons(port);
  serverSocketFd = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);

  if (serverSocketFd < 0) {
    std::cerr << "server socket create error" << std::endl;
    return SOCKET_NOT_CREATE;
  }
  std::cout << "server socket create done!" << std::endl;
  // if success

  if (bind(serverSocketFd, (sockaddr *) &serverAddr, sizeof(sockaddr)) < 0) {
    std::cerr << "server socket bind error " << std::endl;
    close(serverSocketFd);
    return SOCKET_BIND_ERROR;
  }
  std::cout << "server socket listening on " << port << std::endl;

  return SERVER_NO_ERROR;
}

/**
 * register a service ,each service has a unique service name
 * @param methodName service name
 * @param func service registered
 * @return ServerErrorType
 */
JRpc::ServerErrorType JRpc::Server::regService(const MethodName &methodName,
                                               std::function<Json(const Json &)> func) {
  if (serverSocketFd < 0) {
    std::cerr << "Before regService:server socket not create" << std::endl;
    return SOCKET_NOT_CREATE;
  }

  if (this->servicesMap[methodName] != nullptr) {
    std::cerr << "method has beed registered or a reserved method" << std::endl;
    return SERVICE_HAS_REGISTERED;
  }

  servicesMap[methodName] = func;
  return SERVER_NO_ERROR;
}

int JRpc::Server::servicesCount() {
  return this->servicesMap.size();
}

JRpc::Server::~Server() {
  close(serverSocketFd);
}

JRpc::Server::Server(const int port) {
  this->port = port;
  this->serverSocketFd = -1;
  this->version = "1.0";
  init();

  // reserved method for client control server to exit
  regService("_exit", [](const Json &params) {
    Json reply{
        "Server shutdown"
    };
    return reply;
  });

}

/**
 * JRpc server start to serve client
 * @return ServerErrorType
 */
JRpc::ServerErrorType JRpc::Server::serve() {
  if (serverSocketFd < 0) return SOCKET_NOT_CREATE;
  active = true;

  if (listen(serverSocketFd, 5) < 0) {
    return SOCKET_LISTEN_ERROR;
  }

  while (active) {
    // init client tryConnect
    JRpc::JRpcConnection conn;
    bool active = true;
    unsigned int sin_size = sizeof(sockaddr_in);
    memset(conn.buffer, 0, sizeof(conn.buffer));
    conn.socketFd = accept(serverSocketFd, (sockaddr *) &conn.addr, &sin_size);

    if (conn.socketFd < 0) {
      std::cerr << "client socket create failed" << std::endl;
      //sendError(conn.socketFd, JRPC_INVALID_REQUEST, "Not connected");
      close(conn.socketFd);
      continue;
      //return SOCKET_NOT_CREATE;
    }

    std::thread serveAConn(&Server::serveConn, this, conn);
    serveAConn.detach();
  }
  Close();
  return SERVER_NO_ERROR;
}

//    conn.socketFd = accept(serverSocketFd, (sockaddr *) &conn.addr, &sin_size);
//    if (conn.socketFd < 0) {
//      std::cerr << "client socket create failed" << std::endl;
//      sendError(conn.socketFd, JRPC_INVALID_REQUEST, "Not connected");
//      close(conn.socketFd);
//      continue;
//    }
//
//    if (recv(conn.socketFd, conn.buffer, BUFSIZ, 0) == 0) {
//      // no data be sent
//      return SERVER_NO_ERROR;
//    }
//
//    auto ctx = parseRequest(conn);
//    // handle errors
//    if (ctx.errorCode == AS_INFO) {
//      std::cout << "No id attached,consider as an Inform" << std::endl;
//      close(conn.socketFd);
//      continue;
//    }
//
//    if (ctx.errorCode != NO_ERROR) {
//      sendError(conn.socketFd, ctx.errorCode, ctx.errMsg);
//      close(conn.socketFd);
//      continue;
//    }
//
//    auto json = ctx.data;
//    std::function<Json(const Json &)> func;
//    Json result;
//
//    try {
//      func = servicesMap[json["method"]];
//      result = func(json["params"]);
//      sendResponse(conn.socketFd, result, json["id"]);
//      if (json["method"] == "_exit") active = false;
//    } catch (Json::type_error e) {
//      sendError(conn.socketFd, JRPC_INVALID_PARAMS, "Invalid params type or params missing");
//      close(conn.socketFd);
//      continue;
//    }
//    close(conn.socketFd);
//  }
//
//  Close();
//  return SERVER_NO_ERROR;
//
//}

/**
 *  parsing the request and return the context
 * @param conn
 * @return ctx result of parsing
 */
JRpc::JRpcContext JRpc::Server::parseRequest(JRpc::JRpcConnection &conn) {

  JRpcContext ctx;
  auto json = nlohmann::json::parse(conn.buffer, nullptr, false);
  if (json.is_discarded()) {
    ctx.errMsg = "Parsing failed";
    ctx.errorCode = JRPC_PARSE_ERROR;
    ctx.data = nullptr;
    return ctx;
  }

  // other error;
  if (json.size() > 4) {
    ctx.errorCode = JRPC_INVALID_REQUEST;
    ctx.errMsg = "More than three json attributes:method,jsonrpc,params";
    return ctx;
  }

  if (json["id"].is_null() || json["id"].empty()) {
    ctx.errorCode = AS_INFO;
    ctx.errMsg = "As inform,no need to serve";
    return ctx;
  }

  if (json["method"].empty() || json["jsonrpc"].empty()) {
    ctx.errorCode = JRPC_INVALID_REQUEST;
    ctx.errMsg = "Json params missing";
    return ctx;
  }
  if (json["jsonrpc"] != version) {
    ctx.errorCode = JRPC_INVALID_REQUEST;
    ctx.errMsg = "Jsonrpc version not supported";
    return ctx;
  }
  if (servicesMap[json["method"]] == nullptr) {
    ctx.errorCode = JRPC_METHOD_NOT_FOUND;
    ctx.errMsg = "Method not registered";
    return ctx;
  }

  ctx.data = json;

  ctx.errorCode = NO_ERROR;
  ctx.errMsg = "no error";
  return ctx;
}

/**
 * send response to client
 * @param socketFd cli_socket
 * @param result service result
 *
 * @return ServerErrorType
 */
JRpc::ServerErrorType JRpc::Server::sendResponse(int socketFd, Json result, int id) {
  Json reply{
      {"jsonrpc", "1.0"},
      {"result", result},
      //{"error", NO_ERROR},
      {"id", id}
  };
  auto retStr = reply.dump();
  if (send(socketFd, retStr.c_str(), retStr.size(), 0) < 0) {
    return SOCKET_NOT_CREATE;
  }

  return SERVER_NO_ERROR;

}

void JRpc::Server::Close() {
  active = false;
  close(serverSocketFd);
}

/**
 * send a error to client
 * @param socketFd
 * @param code error code
 * @param msg error msg
 * @return ServerErrorType
 */
JRpc::ServerErrorType JRpc::Server::sendError(int socketFd, JRpc::ErrorType code, const std::string msg) {
  Json
      error{
      {"jsonrpc", version},
      {"error", {
          {"code", code},
          {"message", msg}
      }
      },
      {"id", "null"}
  };
  auto errStr = error.dump();
  if (send(socketFd, errStr.c_str(), errStr.size(), 0) < 0) {
    return JRpc::SOCKET_NOT_CREATE;
  }
  return JRpc::SERVER_NO_ERROR;
}

JRpc::ServerErrorType JRpc::Server::serveConn(JRpc::JRpcConnection conn) {
  std::cout << "Serve a connection" << std::endl;

  while (active) {

    if (recv(conn.socketFd, conn.buffer, BUFSIZ, 0) == 0) {
      // no data be sent
      continue;
    }

    auto ctx = parseRequest(conn);
    // handle errors
    if (ctx.errorCode == AS_INFO) {
      std::cout << "No id attached,consider as an Inform" << std::endl;
      continue;
    }

    if (ctx.errorCode != NO_ERROR) {
      sendError(conn.socketFd, ctx.errorCode, ctx.errMsg);
      continue;
    }

    auto json = ctx.data;
    std::function<Json(const Json &)> func;
    Json result;

    try {
      func = servicesMap[json["method"]];
      result = func(json["params"]);
      sendResponse(conn.socketFd, result, json["id"]);
      if (json["method"] == "_disconnect") active = false;
    } catch (Json::type_error e) {
      sendError(conn.socketFd, JRPC_INVALID_PARAMS,
                "Invalid params type or params missing");
      continue;
    }
  }
  close(conn.socketFd);
  return SERVER_NO_ERROR;
}





