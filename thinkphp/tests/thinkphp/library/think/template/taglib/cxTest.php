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
 * 模板测试
 * @author    Haotong Lin <lofanmi@gmail.com>
 */

namespace tests\thinkphp\library\think\tempplate\taglib;

use think\Template;
use think\template\taglib\Cx;

class cxTest extends \PHPUnit_Framework_TestCase
{
    public function testPhp()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{php}echo \$a;{/php}
EOF;
        $data = <<<EOF
<?php echo \$a; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testVolist()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{volist name="list" id="vo" key="key"}

{/volist}
EOF;
        $data = <<<EOF
<?php if(is_array(\$list) || \$list instanceof \\think\\Collection || \$list instanceof \\think\\Paginator): \$key = 0; \$__LIST__ = \$list;if( count(\$__LIST__)==0 ) : echo "" ;else: foreach(\$__LIST__ as \$key=>\$vo): \$mod = (\$key % 2 );++\$key;?>

<?php endforeach; endif; else: echo "" ;endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($data, $content);

        $content = <<<EOF
{volist name="\$list" id="vo" key="key" offset="1" length="3"}
{\$vo}
{/volist}
EOF;

        $template->display($content, ['list' => [1, 2, 3, 4, 5]]);
        $this->expectOutputString('234');
    }

    public function testForeach()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{foreach \$list as \$key=>\$val}

{/foreach}
EOF;
        $data = <<<EOF
<?php foreach(\$list as \$key=>\$val): ?>

<?php endforeach; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{foreach name="list" id="val" key="key" empty="empty"}

{/foreach}
EOF;
        $data = <<<EOF
<?php if(is_array(\$list) || \$list instanceof \\think\\Collection || \$list instanceof \\think\\Paginator): if( count(\$list)==0 ) : echo "empty" ;else: foreach(\$list as \$key=>\$val): ?>

<?php endforeach; endif; else: echo "empty" ;endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{foreach name=":explode(',', '1,2,3,4,5')" id="val" key="key" index="index" mod="2" offset="1" length="3"}
{\$val}
{/foreach}
EOF;
        $template->display($content);
        $this->expectOutputString('234');
    }

    public function testIf()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{if \$var.a==\$var.b}
one
{elseif !empty(\$var.a) /}
two
{else /}
default
{/if}
EOF;
        $data = <<<EOF
<?php if(\$var['a']==\$var['b']): ?>
one
<?php elseif(!empty(\$var['a'])): ?>
two
<?php else: ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testSwitch()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{switch \$var}
{case \$a /}
a
{/case}
{case b|c}
b
{/case}
{case d}
d
{/case}
{default /}
default
{/switch}
EOF;
        $data = <<<EOF
<?php switch(\$var): ?>
<?php case \$a: ?>
a
<?php break; ?>
<?php case "b":case "c": ?>
b
<?php break; ?>
<?php case "d": ?>
d
<?php break; ?>
<?php default: ?>
default
<?php endswitch; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testCompare()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{eq name="\$var.a" value="\$var.b"}
default
{/eq}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] == \$var['b']): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{equal name="\$var.a" value="0"}
default
{/equal}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] == '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{neq name="\$var.a" value="0"}
default
{/neq}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] != '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{notequal name="\$var.a" value="0"}
default
{/notequal}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] != '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{gt name="\$var.a" value="0"}
default
{/gt}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] > '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{egt name="\$var.a" value="0"}
default
{/egt}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] >= '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{lt name="\$var.a" value="0"}
default
{/lt}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] < '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{elt name="\$var.a" value="0"}
default
{/elt}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] <= '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{heq name="\$var.a" value="0"}
default
{/heq}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] === '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{nheq name="\$var.a" value="0"}
default
{/nheq}
EOF;
        $data = <<<EOF
<?php if(\$var['a'] !== '0'): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

    }

    public function testRange()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{in name="var" value="\$value"}
default
{/in}
EOF;
        $data = <<<EOF
<?php if(in_array((\$var), is_array(\$value)?\$value:explode(',',\$value))): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{notin name="var" value="1,2,3"}
default
{/notin}
EOF;
        $data = <<<EOF
<?php if(!in_array((\$var), explode(',',"1,2,3"))): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{between name=":floor(5.1)" value="1,5"}yes{/between}
{notbetween name=":ceil(5.1)" value="1,5"}no{/notbetween}
EOF;
        $template->display($content);
        $this->expectOutputString('yesno');
    }

    public function testPresent()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{present name="var"}
default
{/present}
EOF;
        $data = <<<EOF
<?php if(isset(\$var)): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{notpresent name="var"}
default
{/notpresent}
EOF;
        $data = <<<EOF
<?php if(!isset(\$var)): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testEmpty()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{empty name="var"}
default
{/empty}
EOF;
        $data = <<<EOF
<?php if(empty(\$var) || ((\$var instanceof \\think\\Collection || \$var instanceof \\think\\Paginator ) && \$var->isEmpty())): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{notempty name="var"}
default
{/notempty}
EOF;
        $data = <<<EOF
<?php if(!(empty(\$var) || ((\$var instanceof \\think\\Collection || \$var instanceof \\think\\Paginator ) && \$var->isEmpty()))): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testDefined()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{defined name="URL"}
default
{/defined}
EOF;
        $data = <<<EOF
<?php if(defined("URL")): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{notdefined name="URL"}
default
{/notdefined}
EOF;
        $data = <<<EOF
<?php if(!defined("URL")): ?>
default
<?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testImport()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{load file="base.php" value="\$name.a" /}
EOF;
        $data = <<<EOF
<?php if(isset(\$name['a'])): ?><?php include "base.php"; ?><?php endif; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{js file="base.js" /}
EOF;
        $data = <<<EOF
<script type="text/javascript" src="base.js"></script>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{css file="base.css" /}
EOF;
        $data = <<<EOF
<link rel="stylesheet" type="text/css" href="base.css" />
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testAssign()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{assign name="total" value="0" /}
EOF;
        $data = <<<EOF
<?php \$total = '0'; ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{assign name="total" value=":count(\$list)" /}
EOF;
        $data = <<<EOF
<?php \$total = count(\$list); ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testDefine()
    {
        $template = new template();
        $cx       = new Cx($template);

        $content = <<<EOF
{define name="INFO_NAME" value="test" /}
EOF;
        $data = <<<EOF
<?php define('INFO_NAME', 'test'); ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);

        $content = <<<EOF
{define name="INFO_NAME" value="\$name" /}
EOF;
        $data = <<<EOF
<?php define('INFO_NAME', \$name); ?>
EOF;
        $cx->parseTag($content);
        $this->assertEquals($content, $data);
    }

    public function testFor()
    {
        $template = new template();

        $content = <<<EOF
{for start="1" end=":strlen(1000000000)" comparison="lt" step="1" name="i" }
{\$i}
{/for}
EOF;
        $template->display($content);
        $this->expectOutputString('123456789');
    }
    public function testUrl()
    {
        $template = new template();
        $content  = <<<EOF
{url link="Index/index"  /}
EOF;
        $template->display($content);
        $this->expectOutputString(\think\Url::build('Index/index'));
    }

    public function testFunction()
    {
        $template = new template();
        $data     = [
            'list' => ['language' => 'php', 'version' => ['5.4', '5.5']],
            'a'    => '[',
            'b'    => ']',
        ];

        $content = <<<EOF
{function name="func" vars="\$data" call="\$list" use="&\$a,&\$b"}
{foreach \$data as \$key=>\$val}
{if is_array(\$val)}
{~\$func(\$val)}
{else}
{if !is_numeric(\$key)}
{\$key.':'.\$val.','}
{else}
{\$a.\$val.\$b}
{/if}
{/if}
{/foreach}
{/function}
EOF;
        $template->display($content, $data);
        $this->expectOutputString("language:php,[5.4][5.5]");
    }
}
