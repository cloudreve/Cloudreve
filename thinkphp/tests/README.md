## 测试目录结构

测试文件主要在 tests 文件下面，主要有以下几个文件夹

- conf 测试环境配置文件。
- script 测试环境配置脚本。
- thinkphp 测试用例和相关文件，与项目文件夹结构一致。
- mock.php 测试入口文件。

## 主要测试流程

thinkphp5 的测试的主要流程是跟 thinkphp 的系统流程是相似的，大体的流程为：

1. 引用 mock.php 文件加载框架

2. 根据文件目录，添加测试文件

3. 执行单元测试，输出结果

## 测试举例

例如测试 thinkphp 里的 apc 缓存，将分为以下几个过程：

1. 创建 apcTest.php 文件

该文件应与 apc.php 目录路径 `thinkphp/library/think/cache/driver` 一致，命名空间与目录所在一致，并引用 `PHPUnit_Framework_TestCase`。

  ```php
  <?php

  namespace tests\thinkphp\library\think\cache\driver;

  class apcTest extends \PHPUnit_Framework_TestCase
  {
      //设定基境
      public function setUp()
      {
      }
  }
  ```

2. 编写测试文件

  - 引用 app、config 和 cache

  ```php
  use think\app;
  use think\cache;
  use think\config;
  ```
  - 在 setUp 函数中设定 require 条件

  ```php
  if(!extension_loaded('apc')){
      $this->markTestSkipped('apc扩展不可用！');
  };
  ```

  - 编写测试用例

  *具体写法参照 [PHPUnit 官方文档](https://phpunit.de/manual/4.8/zh_cn/index.html)*

  ```php
  public function testGet()
  {
      App::run();
      $this->assertInstanceOf(
          '\think\cache\driver\Apc',
          Cache::connect(['type' => 'apc', 'expire' => 1])
      );
      $this->assertTrue(Cache::set('key', 'value'));
      $this->assertEquals('value', Cache::get('key'));
      $this->assertTrue(Cache::rm('key'));
      $this->assertFalse(Cache::get('key'));
      $this->assertTrue(Cache::clear('key'));
      Config::reset();
  }
  ```

3. 执行单元测试命令

  在项目根目录执行

  ```bash
  $ phpunit
  ```

  若想看到所有结果，请添加-v参数

  ```bash
  $ phpunit -v
  ```

4. 输出结果

## 相关文档

[各个部分单元测试说明](http://www.kancloud.cn/brother_simon/tp5_test/96971 "各部分单元测试说明")

## 大家一起来

单元测试的内容会跟框架同步，测试内容方方面面，是一个相对复杂的模块，同时也是一个值得重视的部分。希望大家能够多多提出意见，多多参与。如果你有任何问题或想法，可以随时提 issue，我们期待着收到听大家的质疑和讨论。

## 任务进度

单元测试任务进度，请大家认领模块

|模块|认领人|进度|
|---|---|---|
|Base|||
|App|Haotong Lin|√|
|Build|刘志淳||
|Config|Haotong Lin|√|
|Cache|||
|Controller|Haotong Lin|√|
|Cookie|Haotong Lin|√|
|Db|||
|Debug|大漠|√|
|Error|大漠||
|Exception|Haotong Lin|√|
|Hook|流年|√|
|Input|Haotong Lin|√|
|Lang|流年|√|
|Loader|流年||
|Log|||
|Model|||
|Response|大漠|√|
|Route|流年||
|Session|大漠|√|
|Template|oldrind||
|Url|流年||
|View|mahuan||
