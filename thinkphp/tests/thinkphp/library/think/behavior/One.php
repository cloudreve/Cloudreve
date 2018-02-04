<?php
namespace tests\thinkphp\library\think\behavior;

class One
{
    public function run(&$data) {
        $data['id'] = 1;
        return true;
    }

    public function test(&$data) {
        $data['name'] = 'test';
        return false;
    }
}
