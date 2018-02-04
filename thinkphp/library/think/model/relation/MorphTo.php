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

use think\Exception;
use think\Loader;
use think\Model;
use think\model\Relation;

class MorphTo extends Relation
{
    // 多态字段
    protected $morphKey;
    protected $morphType;
    // 多态别名
    protected $alias;
    protected $relation;

    /**
     * 构造函数
     * @access public
     * @param Model  $parent    上级模型对象
     * @param string $morphType 多态字段名
     * @param string $morphKey  外键名
     * @param array  $alias     多态别名定义
     * @param string $relation  关联名
     */
    public function __construct(Model $parent, $morphType, $morphKey, $alias = [], $relation = null)
    {
        $this->parent    = $parent;
        $this->morphType = $morphType;
        $this->morphKey  = $morphKey;
        $this->alias     = $alias;
        $this->relation  = $relation;
    }

    /**
     * 延迟获取关联数据
     * @param string   $subRelation 子关联名
     * @param \Closure $closure     闭包查询条件
     * @return mixed
     */
    public function getRelation($subRelation = '', $closure = null)
    {
        $morphKey  = $this->morphKey;
        $morphType = $this->morphType;
        // 多态模型
        $model = $this->parseModel($this->parent->$morphType);
        // 主键数据
        $pk            = $this->parent->$morphKey;
        $relationModel = (new $model)->relation($subRelation)->find($pk);

        if ($relationModel) {
            $relationModel->setParent(clone $this->parent);
        }
        return $relationModel;
    }

    /**
     * 根据关联条件查询当前模型
     * @access public
     * @param string  $operator 比较操作符
     * @param integer $count    个数
     * @param string  $id       关联表的统计字段
     * @param string  $joinType JOIN类型
     * @return Query
     */
    public function has($operator = '>=', $count = 1, $id = '*', $joinType = 'INNER')
    {
        return $this->parent;
    }

    /**
     * 根据关联条件查询当前模型
     * @access public
     * @param mixed $where 查询条件（数组或者闭包）
     * @return Query
     */
    public function hasWhere($where = [])
    {
        throw new Exception('relation not support: hasWhere');
    }

    /**
     * 解析模型的完整命名空间
     * @access public
     * @param string $model 模型名（或者完整类名）
     * @return string
     */
    protected function parseModel($model)
    {
        if (isset($this->alias[$model])) {
            $model = $this->alias[$model];
        }
        if (false === strpos($model, '\\')) {
            $path = explode('\\', get_class($this->parent));
            array_pop($path);
            array_push($path, Loader::parseName($model, 1));
            $model = implode('\\', $path);
        }
        return $model;
    }

    /**
     * 设置多态别名
     * @access public
     * @param array $alias 别名定义
     * @return $this
     */
    public function setAlias($alias)
    {
        $this->alias = $alias;
        return $this;
    }

    /**
     * 移除关联查询参数
     * @access public
     * @return $this
     */
    public function removeOption()
    {
        return $this;
    }

    /**
     * 预载入关联查询
     * @access public
     * @param array    $resultSet   数据集
     * @param string   $relation    当前关联名
     * @param string   $subRelation 子关联名
     * @param \Closure $closure     闭包
     * @return void
     * @throws Exception
     */
    public function eagerlyResultSet(&$resultSet, $relation, $subRelation, $closure)
    {
        $morphKey  = $this->morphKey;
        $morphType = $this->morphType;
        $range     = [];
        foreach ($resultSet as $result) {
            // 获取关联外键列表
            if (!empty($result->$morphKey)) {
                $range[$result->$morphType][] = $result->$morphKey;
            }
        }

        if (!empty($range)) {
            // 关联属性名
            $attr = Loader::parseName($relation);
            foreach ($range as $key => $val) {
                // 多态类型映射
                $model = $this->parseModel($key);
                $obj   = new $model;
                $pk    = $obj->getPk();
                $list  = $obj->all($val, $subRelation);
                $data  = [];
                foreach ($list as $k => $vo) {
                    $data[$vo->$pk] = $vo;
                }
                foreach ($resultSet as $result) {
                    if ($key == $result->$morphType) {
                        // 关联模型
                        if (!isset($data[$result->$morphKey])) {
                            throw new Exception('relation data not exists :' . $this->model);
                        } else {
                            $relationModel = $data[$result->$morphKey];
                            $relationModel->setParent(clone $result);
                            $relationModel->isUpdate(true);

                            $result->setRelation($attr, $relationModel);
                        }
                    }
                }
            }
        }
    }

    /**
     * 预载入关联查询
     * @access public
     * @param Model    $result      数据对象
     * @param string   $relation    当前关联名
     * @param string   $subRelation 子关联名
     * @param \Closure $closure     闭包
     * @return void
     */
    public function eagerlyResult(&$result, $relation, $subRelation, $closure)
    {
        $morphKey  = $this->morphKey;
        $morphType = $this->morphType;
        // 多态类型映射
        $model = $this->parseModel($result->{$this->morphType});
        $this->eagerlyMorphToOne($model, $relation, $result, $subRelation);
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
     * 多态MorphTo 关联模型预查询
     * @access   public
     * @param object $model       关联模型对象
     * @param string $relation    关联名
     * @param        $result
     * @param string $subRelation 子关联
     * @return void
     */
    protected function eagerlyMorphToOne($model, $relation, &$result, $subRelation = '')
    {
        // 预载入关联查询 支持嵌套预载入
        $pk   = $this->parent->{$this->morphKey};
        $data = (new $model)->with($subRelation)->find($pk);
        if ($data) {
            $data->setParent(clone $result);
            $data->isUpdate(true);
        }
        $result->setRelation(Loader::parseName($relation), $data ?: null);
    }

    /**
     * 添加关联数据
     * @access public
     * @param Model $model       关联模型对象
     * @return Model
     */
    public function associate($model)
    {
        $morphKey  = $this->morphKey;
        $morphType = $this->morphType;
        $pk        = $model->getPk();

        $this->parent->setAttr($morphKey, $model->$pk);
        $this->parent->setAttr($morphType, get_class($model));
        $this->parent->save();

        return $this->parent->setRelation($this->relation, $model);
    }

    /**
     * 注销关联数据
     * @access public
     * @return Model
     */
    public function dissociate()
    {
        $morphKey  = $this->morphKey;
        $morphType = $this->morphType;

        $this->parent->setAttr($morphKey, null);
        $this->parent->setAttr($morphType, null);
        $this->parent->save();

        return $this->parent->setRelation($this->relation, null);
    }

    /**
     * 执行基础查询（进执行一次）
     * @access protected
     * @return void
     */
    protected function baseQuery()
    {}
}
