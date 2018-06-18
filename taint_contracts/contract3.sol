contract contract3 {
    function safeadd(uint arg1, uint arg2) returns (uint) {
        uint res = arg1 + arg2;
        assert(res >= arg1);
        return res;
    }
}