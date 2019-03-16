<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2018 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: zhangyajun <448901948@qq.com>
// +----------------------------------------------------------------------

namespace think;

use ArrayAccess;
use ArrayIterator;
use Countable;
use IteratorAggregate;
use JsonSerializable;

class Collection implements ArrayAccess, Countable, IteratorAggregate, JsonSerializable
{
    /**
     * @var array 数据
     */
    protected $items = [];

    /**
     * Collection constructor.
     * @access public
     * @param  array $items 数据
     */
    public function __construct($items = [])
    {
        $this->items = $this->convertToArray($items);
    }

    /**
     * 创建 Collection 实例
     * @access public
     * @param  array $items 数据
     * @return static
     */
    public static function make($items = [])
    {
        return new static($items);
    }

    /**
     * 判断数据是否为空
     * @access public
     * @return bool
     */
    public function isEmpty()
    {
        return empty($this->items);
    }

    /**
     * 将数据转成数组
     * @access public
     * @return array
     */
    public function toArray()
    {
        return array_map(function ($value) {
            return ($value instanceof Model || $value instanceof self) ?
                $value->toArray() :
                $value;
        }, $this->items);
    }

    /**
     * 获取全部的数据
     * @access public
     * @return array
     */
    public function all()
    {
        return $this->items;
    }

    /**
     * 交换数组中的键和值
     * @access public
     * @return static
     */
    public function flip()
    {
        return new static(array_flip($this->items));
    }

    /**
     * 返回数组中所有的键名组成的新 Collection 实例
     * @access public
     * @return static
     */
    public function keys()
    {
        return new static(array_keys($this->items));
    }

    /**
     * 返回数组中所有的值组成的新 Collection 实例
     * @access public
     * @return static
     */
    public function values()
    {
        return new static(array_values($this->items));
    }

    /**
     * 合并数组并返回一个新的 Collection 实例
     * @access public
     * @param  mixed $items 新的数据
     * @return static
     */
    public function merge($items)
    {
        return new static(array_merge($this->items, $this->convertToArray($items)));
    }

    /**
     * 比较数组，返回差集生成的新 Collection 实例
     * @access public
     * @param  mixed $items 做比较的数据
     * @return static
     */
    public function diff($items)
    {
        return new static(array_diff($this->items, $this->convertToArray($items)));
    }

    /**
     * 比较数组，返回交集组成的 Collection 新实例
     * @access public
     * @param  mixed $items 比较数据
     * @return static
     */
    public function intersect($items)
    {
        return new static(array_intersect($this->items, $this->convertToArray($items)));
    }

    /**
     * 返回并删除数据中的的最后一个元素（出栈）
     * @access public
     * @return mixed
     */
    public function pop()
    {
        return array_pop($this->items);
    }

    /**
     * 返回并删除数据中首个元素
     * @access public
     * @return mixed
     */
    public function shift()
    {
        return array_shift($this->items);
    }

    /**
     * 在数组开头插入一个元素
     * @access public
     * @param mixed $value 值
     * @param mixed $key   键名
     * @return void
     */
    public function unshift($value, $key = null)
    {
        if (is_null($key)) {
            array_unshift($this->items, $value);
        } else {
            $this->items = [$key => $value] + $this->items;
        }
    }

    /**
     * 在数组结尾插入一个元素
     * @access public
     * @param  mixed $value 值
     * @param  mixed $key   键名
     * @return void
     */
    public function push($value, $key = null)
    {
        if (is_null($key)) {
            $this->items[] = $value;
        } else {
            $this->items[$key] = $value;
        }
    }

    /**
     * 通过使用用户自定义函数，以字符串返回数组
     * @access public
     * @param  callable $callback 回调函数
     * @param  mixed    $initial  初始值
     * @return mixed
     */
    public function reduce(callable $callback, $initial = null)
    {
        return array_reduce($this->items, $callback, $initial);
    }

    /**
     * 以相反的顺序创建一个新的 Collection 实例
     * @access public
     * @return static
     */
    public function reverse()
    {
        return new static(array_reverse($this->items));
    }

    /**
     * 把数据分割为新的数组块
     * @access public
     * @param  int  $size         分隔长度
     * @param  bool $preserveKeys 是否保持原数据索引
     * @return static
     */
    public function chunk($size, $preserveKeys = false)
    {
        $chunks = [];

        foreach (array_chunk($this->items, $size, $preserveKeys) as $chunk) {
            $chunks[] = new static($chunk);
        }

        return new static($chunks);
    }

    /**
     * 给数据中的每个元素执行回调
     * @access public
     * @param  callable $callback 回调函数
     * @return $this
     */
    public function each(callable $callback)
    {
        foreach ($this->items as $key => $item) {
            $result = $callback($item, $key);

            if (false === $result) {
                break;
            }

            if (!is_object($item)) {
                $this->items[$key] = $result;
            }
        }

        return $this;
    }

    /**
     * 用回调函数过滤数据中的元素
     * @access public
     * @param callable|null $callback 回调函数
     * @return static
     */
    public function filter(callable $callback = null)
    {
        return new static(array_filter($this->items, $callback ?: null));
    }

    /**
     * 返回数据中指定的一列
     * @access public
     * @param mixed $columnKey 键名
     * @param null  $indexKey  作为索引值的列
     * @return array
     */
    public function column($columnKey, $indexKey = null)
    {
        if (function_exists('array_column')) {
            return array_column($this->items, $columnKey, $indexKey);
        }

        $result = [];
        foreach ($this->items as $row) {
            $key    = $value = null;
            $keySet = $valueSet = false;

            if (null !== $indexKey && array_key_exists($indexKey, $row)) {
                $key    = (string) $row[$indexKey];
                $keySet = true;
            }

            if (null === $columnKey) {
                $valueSet = true;
                $value    = $row;
            } elseif (is_array($row) && array_key_exists($columnKey, $row)) {
                $valueSet = true;
                $value    = $row[$columnKey];
            }

            if ($valueSet) {
                if ($keySet) {
                    $result[$key] = $value;
                } else {
                    $result[] = $value;
                }
            }
        }

        return $result;
    }

    /**
     * 对数据排序，并返回排序后的数据组成的新 Collection 实例
     * @access public
     * @param  callable|null $callback 回调函数
     * @return static
     */
    public function sort(callable $callback = null)
    {
        $items    = $this->items;
        $callback = $callback ?: function ($a, $b) {
            return $a == $b ? 0 : (($a < $b) ? -1 : 1);
        };

        uasort($items, $callback);
        return new static($items);
    }

    /**
     * 将数据打乱后组成新的 Collection 实例
     * @access public
     * @return static
     */
    public function shuffle()
    {
        $items = $this->items;

        shuffle($items);
        return new static($items);
    }

    /**
     * 截取数据并返回新的 Collection 实例
     * @access public
     * @param  int  $offset       起始位置
     * @param  int  $length       截取长度
     * @param  bool $preserveKeys 是否保持原先的键名
     * @return static
     */
    public function slice($offset, $length = null, $preserveKeys = false)
    {
        return new static(array_slice($this->items, $offset, $length, $preserveKeys));
    }

    /**
     * 指定的键是否存在
     * @access public
     * @param  mixed $offset 键名
     * @return bool
     */
    public function offsetExists($offset)
    {
        return array_key_exists($offset, $this->items);
    }

    /**
     * 获取指定键对应的值
     * @access public
     * @param  mixed $offset 键名
     * @return mixed
     */
    public function offsetGet($offset)
    {
        return $this->items[$offset];
    }

    /**
     * 设置键值
     * @access public
     * @param  mixed $offset 键名
     * @param  mixed $value  值
     * @return void
     */
    public function offsetSet($offset, $value)
    {
        if (is_null($offset)) {
            $this->items[] = $value;
        } else {
            $this->items[$offset] = $value;
        }
    }

    /**
     * 删除指定键值
     * @access public
     * @param  mixed $offset 键名
     * @return void
     */
    public function offsetUnset($offset)
    {
        unset($this->items[$offset]);
    }

    /**
     * 统计数据的个数
     * @access public
     * @return int
     */
    public function count()
    {
        return count($this->items);
    }

    /**
     * 获取数据的迭代器
     * @access public
     * @return ArrayIterator
     */
    public function getIterator()
    {
        return new ArrayIterator($this->items);
    }

    /**
     * 将数据反序列化成数组
     * @access public
     * @return array
     */
    public function jsonSerialize()
    {
        return $this->toArray();
    }

    /**
     * 转换当前数据集为 JSON 字符串
     * @access public
     * @param  integer $options json 参数
     * @return string
     */
    public function toJson($options = JSON_UNESCAPED_UNICODE)
    {
        return json_encode($this->toArray(), $options);
    }

    /**
     * 将数据转换成字符串
     * @access public
     * @return string
     */
    public function __toString()
    {
        return $this->toJson();
    }

    /**
     * 将数据转换成数组
     * @access protected
     * @param  mixed $items 数据
     * @return array
     */
    protected function convertToArray($items)
    {
        return $items instanceof self ? $items->all() : (array) $items;
    }
}
