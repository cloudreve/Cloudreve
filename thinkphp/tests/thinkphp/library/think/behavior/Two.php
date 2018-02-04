<?php
namespace tests\thinkphp\library\think\behavior;

class Two
{
    public function run(&$data) {
        $data['id'] = 2;
    }
}
