<?php
namespace tests\thinkphp\library\think\behavior;

class Three
{
    public function run(&$data) {
        $data['id'] = 3;
    }
}
