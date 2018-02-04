<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------

/**
 * Db类测试
 * @author: 刘志淳 <chun@engineer.com>
 */

namespace tests\thinkphp\library\think;

use think\Db;

class dbTest extends \PHPUnit_Framework_TestCase
{
    // 获取测试数据库配置
    private function getConfig()
    {
        return [
            // 数据库类型
            'type'           => 'mysql',
            // 服务器地址
            'hostname'       => '127.0.0.1',
            // 数据库名
            'database'       => 'test',
            // 用户名
            'username'       => 'root',
            // 密码
            'password'       => '',
            // 端口
            'hostport'       => '',
            // 连接dsn
            'dsn'            => '',
            // 数据库连接参数
            'params'         => [],
            // 数据库编码默认采用utf8
            'charset'        => 'utf8',
            // 数据库表前缀
            'prefix'         => 'tp_',
            // 数据库调试模式
            'debug'          => true,
            // 数据库部署方式:0 集中式(单一服务器),1 分布式(主从服务器)
            'deploy'         => 0,
            // 数据库读写是否分离 主从式有效
            'rw_separate'    => false,
            // 读写分离后 主服务器数量
            'master_num'     => 1,
            // 指定从服务器序号
            'slave_no'       => '',
            // 是否严格检查字段是否存在
            'fields_strict'  => true,
            // 数据集返回类型 array 数组 collection Collection对象
            'resultset_type' => 'array',
            // 是否自动写入时间戳字段
            'auto_timestamp' => false,
            // 是否需要进行SQL性能分析
            'sql_explain'    => false,
        ];
    }

    // 获取创建数据库 SQL
    private function getCreateTableSql()
    {
        $sql[] = <<<EOF
DROP TABLE IF EXISTS `tp_user`;
EOF;
        $sql[] = <<<EOF
CREATE TABLE `tp_user` (
  `id` int(10) unsigned NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `username` char(40) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` char(40) NOT NULL DEFAULT '' COMMENT '密码',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `create_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间'
) ENGINE=MyISAM AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='会员表';
EOF;
        $sql[] = <<<EOF
ALTER TABLE `tp_user` ADD INDEX(`create_time`);
EOF;
        $sql[] = <<<EOF
DROP TABLE IF EXISTS `tp_order`;
EOF;
        $sql[] = <<<EOF
CREATE TABLE `tp_order` (
  `id` int(10) unsigned NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
  `sn` char(20) NOT NULL DEFAULT '' COMMENT '订单号',
  `amount` decimal(10,2) unsigned NOT NULL DEFAULT '0' COMMENT '金额',
  `freight_fee` decimal(10,2) unsigned NOT NULL DEFAULT '0' COMMENT '运费',
  `address_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '地址id',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `create_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间'
) ENGINE=MyISAM AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='订单表';
EOF;
        $sql[] = <<<EOF
DROP TABLE IF EXISTS `tp_user_address`;
EOF;
        $sql[] = <<<EOF
CREATE TABLE `tp_user_address` (
  `id` int(10) unsigned NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
  `consignee` varchar(60) NOT NULL DEFAULT '' COMMENT '收货人',
  `area_info` varchar(50) NOT NULL DEFAULT '' COMMENT '地区信息',
  `city_id` smallint(5) unsigned NOT NULL DEFAULT '0' COMMENT '城市id',
  `area_id` smallint(5) unsigned NOT NULL DEFAULT '0' COMMENT '地区id',
  `address` varchar(120) NOT NULL DEFAULT '' COMMENT '地址',
  `tel` varchar(60) NOT NULL DEFAULT '' COMMENT '电话',
  `mobile` varchar(60) NOT NULL DEFAULT '' COMMENT '手机',
  `isdefault` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '是否默认'
) ENGINE=MyISAM AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='地址表';
EOF;
        $sql[] = <<<EOF
DROP TABLE IF EXISTS `tp_role_user`;
EOF;
        $sql[] = <<<EOF
CREATE TABLE `tp_role_user` (
  `role_id` smallint(5) unsigned NOT NULL,
  `user_id` int(10) unsigned NOT NULL,
  `remark` varchar(250) NOT NULL DEFAULT '',
  PRIMARY KEY (`role_id`,`user_id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
EOF;

        return $sql;
    }

    public function testConnect()
    {
        $config = $this->getConfig();
        $result = Db::connect($config)->execute('show databases');
        $this->assertNotEmpty($result);
    }

    public function testExecute()
    {
        $config = $this->getConfig();
        $sql    = $this->getCreateTableSql();
        foreach ($sql as $one) {
            Db::connect($config)->execute($one);
        }
        $tableNum = Db::connect($config)->execute("show tables;");
        $this->assertEquals(4, $tableNum);
    }

    public function testQuery()
    {
        $config = $this->getConfig();
        $sql    = $this->getCreateTableSql();
        Db::connect($config)->batchQuery($sql);

        $tableQueryResult = Db::connect($config)->query("show tables;");

        $this->assertTrue(is_array($tableQueryResult));

        $tableNum = count($tableQueryResult);
        $this->assertEquals(4, $tableNum);
    }

    public function testBatchQuery()
    {
        $config = $this->getConfig();
        $sql    = $this->getCreateTableSql();
        Db::connect($config)->batchQuery($sql);

        $tableNum = Db::connect($config)->execute("show tables;");
        $this->assertEquals(4, $tableNum);
    }

    public function testTable()
    {
        $config    = $this->getConfig();
        $tableName = 'tp_user';
        $result    = Db::connect($config)->table($tableName);
        $this->assertEquals($tableName, $result->getOptions()['table']);
    }

    public function testName()
    {
        $config    = $this->getConfig();
        $tableName = 'user';
        $result    = Db::connect($config)->name($tableName);
        $this->assertEquals($config['prefix'] . $tableName, $result->getTable());
    }

    public function testInsert()
    {
        $config = $this->getConfig();
        $data   = [
            'username'    => 'chunice',
            'password'    => md5('chunice'),
            'status'      => 1,
            'create_time' => time(),
        ];
        $result = Db::connect($config)->name('user')->insert($data);
        $this->assertEquals(1, $result);
    }

    public function testUpdate()
    {
        $config = $this->getConfig();
        $data   = [
            'username'    => 'chunice_update',
            'password'    => md5('chunice'),
            'status'      => 1,
            'create_time' => time(),
        ];
        $result = Db::connect($config)->name('user')->where('username', 'chunice')->update($data);
        $this->assertEquals(1, $result);
    }

    public function testFind()
    {
        $config   = $this->getConfig();
        $mustFind = Db::connect($config)->name('user')->where('username', 'chunice_update')->find();
        $this->assertNotEmpty($mustFind);
        $mustNotFind = Db::connect($config)->name('user')->where('username', 'chunice')->find();
        $this->assertEmpty($mustNotFind);
    }

    public function testInsertAll()
    {
        $config = $this->getConfig();

        $data = [
            ['username' => 'foo', 'password' => md5('foo'), 'status' => 1, 'create_time' => time()],
            ['username' => 'bar', 'password' => md5('bar'), 'status' => 1, 'create_time' => time()],
        ];

        $insertNum = Db::connect($config)->name('user')->insertAll($data);
        $this->assertEquals(count($data), $insertNum);
    }

    public function testSelect()
    {
        $config    = $this->getConfig();
        $mustFound = Db::connect($config)->name('user')->where('status', 1)->select();
        $this->assertNotEmpty($mustFound);
        $mustNotFound = Db::connect($config)->name('user')->where('status', 0)->select();
        $this->assertEmpty($mustNotFound);
    }

    public function testValue()
    {
        $config   = $this->getConfig();
        $username = Db::connect($config)->name('user')->where('id', 1)->value('username');
        $this->assertEquals('chunice_update', $username);
        $usernameNull = Db::connect($config)->name('user')->where('id', 0)->value('username');
        $this->assertEmpty($usernameNull);
    }

    public function testColumn()
    {
        $config   = $this->getConfig();
        $username = Db::connect($config)->name('user')->where('status', 1)->column('username');
        $this->assertNotEmpty($username);
        $usernameNull = Db::connect($config)->name('user')->where('status', 0)->column('username');
        $this->assertEmpty($usernameNull);

    }

    public function testInsertGetId()
    {
        $config = $this->getConfig();
        $id     = Db::connect($config)->name('user')->order('id', 'desc')->value('id');

        $data = [
            'username'    => uniqid(),
            'password'    => md5('chunice'),
            'status'      => 1,
            'create_time' => time(),
        ];
        $lastId = Db::connect($config)->name('user')->insertGetId($data);
        $this->assertEquals($id + 1, $lastId);

    }

    public function testGetLastInsId()
    {
        $config = $this->getConfig();
        $data   = [
            'username'    => uniqid(),
            'password'    => md5('chunice'),
            'status'      => 1,
            'create_time' => time(),
        ];
        $lastId = Db::connect($config)->name('user')->insertGetId($data);

        $lastInsId = Db::connect($config)->name('user')->getLastInsID();
        $this->assertEquals($lastId, $lastInsId);
    }

    public function testSetField()
    {
        $config = $this->getConfig();

        $setFieldNum = Db::connect($config)->name('user')->where('id', 1)->setField('username', 'chunice_setField');
        $this->assertEquals(1, $setFieldNum);

        $setFieldNum = Db::connect($config)->name('user')->where('id', 1)->setField('username', 'chunice_setField');
        $this->assertEquals(0, $setFieldNum);
    }

    public function testSetInc()
    {
        $config           = $this->getConfig();
        $originCreateTime = Db::connect($config)->name('user')->where('id', 1)->value('create_time');
        Db::connect($config)->name('user')->where('id', 1)->setInc('create_time');
        $newCreateTime = Db::connect($config)->name('user')->where('id', 1)->value('create_time');
        $this->assertEquals($originCreateTime + 1, $newCreateTime);

    }

    public function testSetDec()
    {
        $config           = $this->getConfig();
        $originCreateTime = Db::connect($config)->name('user')->where('id', 1)->value('create_time');
        Db::connect($config)->name('user')->where('id', 1)->setDec('create_time');
        $newCreateTime = Db::connect($config)->name('user')->where('id', 1)->value('create_time');
        $this->assertEquals($originCreateTime - 1, $newCreateTime);
    }

    public function testDelete()
    {
        $config = $this->getConfig();
        Db::connect($config)->name('user')->where('id', 1)->delete();
        $result = Db::connect($config)->name('user')->where('id', 1)->find();
        $this->assertEmpty($result);
    }

    public function testChunk()
    {
        // todo 暂未想到测试方法
    }

    public function testCache()
    {
        $config = $this->getConfig();
        $result = Db::connect($config)->name('user')->where('id', 1)->cache('key', 60)->find();
        $cache  = \think\Cache::get('key');
        $this->assertEquals($result, $cache);

        $updateCache = Db::connect($config)->name('user')->cache('key')->find(1);
        $this->assertEquals($cache, $updateCache);
    }

}
