# json-rpc
## description
- a simple json-rpc framework
- provide a Server and Client interface for RPC.
- based on socket API and [json](https://github.com/nlohmann/json).
- use std::thread for multi-thread serving connection,with a blocking client
## usage
**JRpc::Server**
```c++
  JRpc::Server server(PORT);

  server.regService("ECHO", [](const Json &params) {
                      return params;
                    }
  );

  server.regService("SUM", [](const Json &params) {
                      auto one=params[0].get<int>();
                      auto two=params[0].get<int>();
                      Json reply{
                        one+two
                      };
                      return reply;
                    }
  );
  server.regService("MUL", [](const Json &params) {
                      auto one=params[0].get<int>();
                      auto two=params[0].get<int>();
                      Json reply{
                          one*two
                      };
                      return reply;
                    }
  );

  server.serve();
  server.Close();
```
**JRpc::Client**
```c++
  auto cli = JRpc::Client(Addr, PORT);s
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
```
## feature
- Use [json](https://github.com/nlohmann/json) as the carrier of params and reply.
- use socket for network communication
## reference
- [jsor-rpc](https://github.com/hmng/jsonrpc-c)
- [specification_v1](https://www.jsonrpc.org/specification_v1)
- [specification_v2](https://www.jsonrpc.org/specification_v2)


