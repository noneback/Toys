//
// Created by chenlan on 2020/12/18.
//

#ifndef JSON_RPC__SHARED_H_
#define JSON_RPC__SHARED_H_
#include <string>
#include "json.h"
using Json=nlohmann::json;
namespace JRpc{

struct Reply{
  std::string id;
  Json result;
  std::string jsonrpc;
};

struct Request{
  std::string method;
  std::string id;
  std::string jsonrpc;
  Json params;
};

struct Error{
  Json error;
  std::string jsonrpc;
  std::string id;
};

}

#endif //JSON_RPC__SHARED_H_
