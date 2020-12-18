#include <iostream>
#include <vector>
#include "Server.h"
#include "Client.h"

int main() {
  auto cli = JRpc::Client("127.0.0.1", 9939);
  if (!cli.tryConnect()) {
    std::cout << "tryConnect failed" << std::endl;
  }
  Json params{
      1
  };
  Json params1{
      1,2
  };
  Json params2{
      "1",2
  };
  Json reply;
  std::cout<<params<<std::endl;
  cli.call("ECHO", params, reply);
  std::cout << reply << std::endl;
  Json reply1;
  std::cout<<params1<<std::endl;
  cli.call("MULa", params1, reply1);
  std::cout << reply1 << std::endl;
  Json reply2;
  std::cout<<params2<<std::endl;
  cli.call("SUM", params2, reply2);
  std::cout << reply2 << std::endl;
  cli.disconnect();



//  JRpc::Server server(9939);
//
//  server.regService("ECHO", [](const Json &params) {
//                      return params;
//                    }
//  );
//
//  server.regService("SUM", [](const Json &params) {
//                      auto one=params[0].get<int>();
//                      auto two=params[0].get<int>();
//                      Json reply{
//                        one+two
//                      };
//                      return reply;
//                    }
//  );
//  server.regService("MUL", [](const Json &params) {
//                      auto one=params[0].get<int>();
//                      auto two=params[0].get<int>();
//                      Json reply{
//                          one*two
//                      };
//                      return reply;
//                    }
//  );
//
//  server.serve();
  //server.Close();
  return 0;
}
