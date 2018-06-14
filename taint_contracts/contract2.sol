contract contract2 {
    function f(uint n) returns (uint) {
        return n;
    }

    function callf(uint a) returns (uint) {
        return this.f(a);
    }
}
