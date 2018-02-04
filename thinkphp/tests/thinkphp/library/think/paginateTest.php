<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK IT ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006-2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: yunwuxin <448901948@qq.com>
// +----------------------------------------------------------------------

namespace tests\thinkphp\library\think;

use think\paginator\driver\Bootstrap;

class paginateTest extends \PHPUnit_Framework_TestCase
{
    public function testPaginatorInfo()
    {
        $p = Bootstrap::make($array = ['item3', 'item4'], 2, 2, 4);

        $this->assertEquals(4, $p->total());

        $this->assertEquals(2, $p->listRows());

        $this->assertEquals(2, $p->currentPage());

        $p2 = Bootstrap::make($array2 = ['item3', 'item4'], 2, 2, 2);
        $this->assertEquals(1, $p2->currentPage());
    }

    public function testPaginatorRender()
    {
        $p      = Bootstrap::make($array = ['item3', 'item4'], 2, 2, 100);
        $render = '<ul class="pagination"><li><a href="/?page=1">&laquo;</a></li> <li><a href="/?page=1">1</a></li><li class="active"><span>2</span></li><li><a href="/?page=3">3</a></li><li><a href="/?page=4">4</a></li><li><a href="/?page=5">5</a></li><li><a href="/?page=6">6</a></li><li><a href="/?page=7">7</a></li><li><a href="/?page=8">8</a></li><li class="disabled"><span>...</span></li><li><a href="/?page=49">49</a></li><li><a href="/?page=50">50</a></li> <li><a href="/?page=3">&raquo;</a></li></ul>';

        $this->assertEquals($render, $p->render());
    }

}
