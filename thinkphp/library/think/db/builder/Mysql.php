<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2018 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------

namespace think\db\builder;

use think\db\Builder;
use think\Exception;

/**
 * mysql数据库驱动
 */
class Mysql extends Builder
{

    protected $insertAllSql = '%INSERT% INTO %TABLE% (%FIELD%) VALUES %DATA% %COMMENT%';
    protected $updateSql    = 'UPDATE %TABLE% %JOIN% SET %SET% %WHERE% %ORDER%%LIMIT% %LOCK%%COMMENT%';

    /**
     * 生成insertall SQL
     * @access public
     * @param array     $dataSet 数据集
     * @param array     $options 表达式
     * @param bool      $replace 是否replace
     * @return string
     * @throws Exception
     */
    public function insertAll($dataSet, $options = [], $replace = false)
    {
        // 获取合法的字段
        if ('*' == $options['field']) {
            $fields = array_keys($this->query->getFieldsType($options['table']));
        } else {
            $fields = $options['field'];
        }

        foreach ($dataSet as $data) {
            foreach ($data as $key => $val) {
                if (!in_array($key, $fields, true)) {
                    if ($options['strict']) {
                        throw new Exception('fields not exists:[' . $key . ']');
                    }
                    unset($data[$key]);
                } elseif (is_null($val)) {
                    $data[$key] = 'NULL';
                } elseif (is_scalar($val)) {
                    $data[$key] = $this->parseValue($val, $key);
                } elseif (is_object($val) && method_exists($val, '__toString')) {
                    // 对象数据写入
                    $data[$key] = $val->__toString();
                } else {
                    // 过滤掉非标量数据
                    unset($data[$key]);
                }
            }
            $value    = array_values($data);
            $values[] = '( ' . implode(',', $value) . ' )';

            if (!isset($insertFields)) {
                $insertFields = array_map([$this, 'parseKey'], array_keys($data));
            }
        }

        return str_replace(
            ['%INSERT%', '%TABLE%', '%FIELD%', '%DATA%', '%COMMENT%'],
            [
                $replace ? 'REPLACE' : 'INSERT',
                $this->parseTable($options['table'], $options),
                implode(' , ', $insertFields),
                implode(' , ', $values),
                $this->parseComment($options['comment']),
            ], $this->insertAllSql);
    }

    /**
     * 字段和表名处理
     * @access protected
     * @param mixed  $key
     * @param array  $options
     * @return string
     */
    protected function parseKey($key, $options = [], $strict = false)
    {
        if (is_numeric($key)) {
            return $key;
        } elseif ($key instanceof Expression) {
            return $key->getValue();
        }

        $key = trim($key);
        if (strpos($key, '$.') && false === strpos($key, '(')) {
            // JSON字段支持
            list($field, $name) = explode('$.', $key);
            return 'json_extract(' . $field . ', \'$.' . $name . '\')';
        } elseif (strpos($key, '.') && !preg_match('/[,\'\"\(\)`\s]/', $key)) {
            list($table, $key) = explode('.', $key, 2);
            if ('__TABLE__' == $table) {
                $table = $this->query->getTable();
            }
            if (isset($options['alias'][$table])) {
                $table = $options['alias'][$table];
            }
        }

        if ($strict && !preg_match('/^[\w\.\*]+$/', $key)) {
            throw new Exception('not support data:' . $key);
        }
        if ('*' != $key && ($strict || !preg_match('/[,\'\"\*\(\)`.\s]/', $key))) {
            $key = '`' . $key . '`';
        }
        if (isset($table)) {
            if (strpos($table, '.')) {
                $table = str_replace('.', '`.`', $table);
            }
            $key = '`' . $table . '`.' . $key;
        }
        return $key;
    }

    /**
     * 随机排序
     * @access protected
     * @return string
     */
    protected function parseRand()
    {
        return 'rand()';
    }

}
