<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2017 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------

namespace think\model\relation;

use think\db\Query;
use think\Exception;
use think\Loader;
use think\Model;
use think\model\Relation;

/**
 * Class OneToOne
 * @package think\model\relation
 *
 */
abstract class OneToOne extends Relation
{
    // 预载入方式 0 -JOIN 1 -IN
    protected $eagerlyType = 1;
    // 当前关联的JOIN类型
    protected $joinType;
    // 要绑定的属性
    protected $bindAttr = [];
    // 关联方法名
    protected $relation;

    /**
     * 设置join类型
     * @access public
     * @param string $type JOIN类型
     * @return $this
     */
    public function joinType($type)
    {
        $this->joinType = $type;
        return $this;
    }

    /**
     * 预载入关联查询（JOIN方式）
     * @access public
     * @param Query    $query       查询对象
     * @param string   $relation    关联名
     * @param string   $subRelation 子关联
     * @param \Closure $closure     闭包条件
     * @param bool     $first
     * @return void
     */
    public function eagerly(Query $query, $relation, $subRelation, $closure, $first)
    {
        $name  = Loader::parseName(basename(str_replace('\\', '/', $query->getModel())));
        $alias = $name;
        if ($first) {
            $table = $query->getTable();
            $query->table([$table => $alias]);
            if ($query->getOptions('field')) {
                $field = $query->getOptions('field');
                $query->removeOption('field');
            } else {
                $field = true;
            }
            $query->field($field, false, $table, $alias);
            $field = null;
        }

        // 预载入封装
        $joinTable = $this->query->getTable();
        $joinAlias = $relation;
        $query->via($joinAlias);

        if ($this instanceof BelongsTo) {
            $query->join($joinTable . ' ' . $joinAlias, $alias . '.' . $this->foreignKey . '=' . $joinAlias . '.' . $this->localKey, $this->joinType);
        } else {
            $query->join($joinTable . ' ' . $joinAlias, $alias . '.' . $this->localKey . '=' . $joinAlias . '.' . $this->foreignKey, $this->joinType);
        }

        if ($closure) {
            // 执行闭包查询
            call_user_func_array($closure, [ & $query]);
            // 使用withField指定获取关联的字段，如
            // $query->where(['id'=>1])->withField('id,name');
            if ($query->getOptions('with_field')) {
                $field = $query->getOptions('with_field');
                $query->removeOption('with_field');
            }
        } elseif (isset($this->option['field'])) {
            $field = $this->option['field'];
        }
        $query->field(isset($field) ? $field : true, false, $joinTable, $joinAlias, $relation . '__');
    }

    /**
     *  预载入关联查询（数据集）
     * @param array    $resultSet
     * @param string   $relation
     * @param string   $subRelation
     * @param \Closure $closure
     * @return mixed
     */
    abstract protected function eagerlySet(&$resultSet, $relation, $subRelation, $closure);

    /**
     * 预载入关联查询（数据）
     * @param Model    $result
     * @param string   $relation
     * @param string   $subRelation
     * @param \Closure $closure
     * @return mixed
     */
    abstract protected function eagerlyOne(&$result, $relation, $subRelation, $closure);

    /**
     * 预载入关联查询（数据集）
     * @access public
     * @param array    $resultSet   数据集
     * @param string   $relation    当前关联名
     * @param string   $subRelation 子关联名
     * @param \Closure $closure     闭包
     * @return void
     */
    public function eagerlyResultSet(&$resultSet, $relation, $subRelation, $closure)
    {
        if (1 == $this->eagerlyType) {
            // IN查询
            $this->eagerlySet($resultSet, $relation, $subRelation, $closure);
        } else {
            // 模型关联组装
            foreach ($resultSet as $result) {
                $this->match($this->model, $relation, $result);
            }
        }
    }

    /**
     * 预载入关联查询（数据）
     * @access public
     * @param Model    $result      数据对象
     * @param string   $relation    当前关联名
     * @param string   $subRelation 子关联名
     * @param \Closure $closure     闭包
     * @return void
     */
    public function eagerlyResult(&$result, $relation, $subRelation, $closure)
    {
        if (1 == $this->eagerlyType) {
            // IN查询
            $this->eagerlyOne($result, $relation, $subRelation, $closure);
        } else {
            // 模型关联组装
            $this->match($this->model, $relation, $result);
        }
    }

    /**
     * 保存（新增）当前关联数据对象
     * @access public
     * @param mixed $data 数据 可以使用数组 关联模型对象 和 关联对象的主键
     * @return Model|false
     */
    public function save($data)
    {
        if ($data instanceof Model) {
            $data = $data->getData();
        }
        $model = new $this->model;
        // 保存关联表数据
        $data[$this->foreignKey] = $this->parent->{$this->localKey};
        return $model->save($data) ? $model : false;
    }

    /**
     * 设置预载入方式
     * @access public
     * @param integer $type 预载入方式 0 JOIN查询 1 IN查询
     * @return $this
     */
    public function setEagerlyType($type)
    {
        $this->eagerlyType = $type;
        return $this;
    }

    /**
     * 获取预载入方式
     * @access public
     * @return integer
     */
    public function getEagerlyType()
    {
        return $this->eagerlyType;
    }

    /**
     * 绑定关联表的属性到父模型属性
     * @access public
     * @param mixed $attr 要绑定的属性列表
     * @return $this
     */
    public function bind($attr)
    {
        if (is_string($attr)) {
            $attr = explode(',', $attr);
        }
        $this->bindAttr = $attr;
        return $this;
    }

    /**
     * 关联统计
     * @access public
     * @param Model    $result  数据对象
     * @param \Closure $closure 闭包
     * @return integer
     */
    public function relationCount($result, $closure)
    {
    }

    /**
     * 一对一 关联模型预查询拼装
     * @access public
     * @param string $model    模型名称
     * @param string $relation 关联名
     * @param Model  $result   模型对象实例
     * @return void
     */
    protected function match($model, $relation, &$result)
    {
        // 重新组装模型数据
        foreach ($result->getData() as $key => $val) {
            if (strpos($key, '__')) {
                list($name, $attr) = explode('__', $key, 2);
                if ($name == $relation) {
                    $list[$name][$attr] = $val;
                    unset($result->$key);
                }
            }
        }

        if (isset($list[$relation])) {
            $relationModel = new $model($list[$relation]);
            $relationModel->setParent(clone $result);
            $relationModel->isUpdate(true);

            if (!empty($this->bindAttr)) {
                $this->bindAttr($relationModel, $result, $this->bindAttr);
            }
        } else {
            $relationModel = null;
        }
        $result->setRelation(Loader::parseName($relation), $relationModel);
    }

    /**
     * 绑定关联属性到父模型
     * @access protected
     * @param Model $model    关联模型对象
     * @param Model $result   父模型对象
     * @param array $bindAttr 绑定属性
     * @return void
     * @throws Exception
     */
    protected function bindAttr($model, &$result, $bindAttr)
    {
        foreach ($bindAttr as $key => $attr) {
            $key = is_numeric($key) ? $attr : $key;
            if (isset($result->$key)) {
                throw new Exception('bind attr has exists:' . $key);
            } else {
                $result->setAttr($key, $model->$attr);
            }
        }
    }

    /**
     * 一对一 关联模型预查询（IN方式）
     * @access public
     * @param object        $model       关联模型对象
     * @param array         $where       关联预查询条件
     * @param string        $key         关联键名
     * @param string        $relation    关联名
     * @param string        $subRelation 子关联
     * @param bool|\Closure $closure
     * @return array
     */
    protected function eagerlyWhere($model, $where, $key, $relation, $subRelation = '', $closure = false)
    {
        // 预载入关联查询 支持嵌套预载入
        if ($closure) {
            call_user_func_array($closure, [ & $model]);
            if ($field = $model->getOptions('with_field')) {
                $model->field($field)->removeOption('with_field');
            }
        }
        $list = $model->where($where)->with($subRelation)->select();

        // 组装模型数据
        $data = [];
        foreach ($list as $set) {
            $data[$set->$key] = $set;
        }
        return $data;
    }

    /**
     * 执行基础查询（进执行一次）
     * @access protected
     * @return void
     */
    protected function baseQuery()
    {}
}
