contract Agent {
    uint public count = 0;
    address public call_contract_addr;
    bytes public call_msg_data;
    bool public turnoff = true;
    bool public hasValue = false;
    uint public sendCount = 0;
    uint public sendFailedCount = 0;

    // 回退函数，允许接收以太币
    function() payable {
        if (turnoff) {
            count++;
            // 在执行 "call_contract_addr.call.." 语句之前将 turnoff 设置为 false。
            // 因为我们只需要测试一次重入攻击，
            // 多次测试重入攻击是没有必要的。
            turnoff = false;
            call_contract_addr.call(call_msg_data);
        } else {
            turnoff = true;
        }
    }

    // 构造函数
    function Agent() {

    }

    // 获取合约地址
    function getContractAddr() returns (address addr) {
        return call_contract_addr;
    }

    // 获取调用的消息数据
    function getCallMsgData() returns (bytes msg_data) {
        return call_msg_data;
    }

    // 无附加以太币的合约调用
    function AgentCallWithoutValue(address contract_addr, bytes msg_data) {
        hasValue = false;
        call_contract_addr = contract_addr;
        call_msg_data = msg_data;
        contract_addr.call(msg_data);
    }

    // 附加以太币的合约调用
    function AgentCallWithValue(address contract_addr, bytes msg_data) payable {
        hasValue = true;
        uint msg_value = msg.value;
        call_contract_addr = contract_addr;
        call_msg_data = msg_data;
        contract_addr.call.value(msg_value)(msg_data);
    }

    // 发送以太币到指定合约地址
    function AgentSend(address contract_addr) payable {
        sendCount++;
        if (!contract_addr.send(msg.value))
            sendFailedCount++;
    }
}
