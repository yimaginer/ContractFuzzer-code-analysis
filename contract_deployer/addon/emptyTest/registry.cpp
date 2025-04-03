#include <node.h>
void RegisterModule(v8::Handle<v8::Object> target) {
		// 注册模块功能，负责导出接口到node.js
}
// 注册模块名称，编译后，模块将编译成modulename.node文件
// 当你需要修改模块名字的时候，需要修改 binding.gyp("target_name") 和此处
NODE_MODULE(registry, RegisterModule);

// node.h 是开发 Node.js 原生扩展模块的核心头文件，提供了与 Node.js 和 V8 引擎交互的接口。
// 在你的代码中，它用于注册模块（registry），并将 C++ 的功能导出到 Node.js 环境中。