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

use think\Collection;
use think\db\Query;
use think\Exception;
use think\Loader;
use think\Model;
use think\model\Pivot;
use think\model\Relation;
use think\Paginator;

class BelongsToMany extends Relation
{
    // 中间表表名
    protected $middle;
    // 中间表模型名称
    protected $pivotName;
    // 中间表模型对象
    protected $pivot;

    /**
     * 构造函数
     * @access public
     * @param Model  $parent     上级模型对象
     * @param string $model      模型名
     * @param string $table      中间表名
     * @param string $foreignKey 关联模型外键
     * @param string $localKey   当前模型关联键
     */
    public function __construct(Model $parent, $model, $table, $foreignKey, $localKey)
    {
        $this->parent     = $parent;
        $this->model      = $model;
        $this->foreignKey = $foreignKey;
        $this->localKey   = $localKey;
        if (false !== strpos($table, '\\')) {
            $this->pivotName = $table;
            $this->middle    = basename(str_replace('\\', '/', $table));
        } else {
            $this->middle = $table;
        }
        $this->query = (new $model)->db();
        $this->pivot = $this->newPivot();
    }

    /**
     * 设置中间表模型
     * @param $pivot
     * @return $this
     */
    public function pivot($pivot)
    {
        $this->pivotName = $pivot;
        return $this;
    }

    /**
     * 实例化中间表模型
     * @param $data
     * @return mixed
     */
    protected function newPivot($data = [])
    {
        $pivot = $this->pivotName ?: '\\think\\model\\Pivot';
        return new $pivot($this->parent, $data, $this->middle);
    }

    /**
     * 合成中间表模型
     * @param array|Collection|Paginator $models
     */
    protected function hydratePivot($models)
    {
        foreach ($models as $model) {
            $pivot = [];
            foreach ($model->getData() as $key => $val) {
                if (strpos($key, '__')) {
                    list($name, $attr) = explode('__', $key, 2);
                    if ('pivot' == $name) {
                        $pivot[$attr] = $val;
                        unset($model->$key);
                    }
                }
            }
            $model->setRelation('pivot', $this->newPivot($pivot));
        }
    }

    /**
     * 创建关联查询Query对象
     * @return Query
     */
    protected function buildQuery()
    {
        $foreignKey = $this->foreignKey;
        $localKey   = $this->localKey;
        $middle     = $this->middle;
        // 关联查询
        $pk                              = $this->parent->getPk();
        $condition['pivot.' . $localKey] = $this->parent->$pk;
        return $this->belongsToManyQuery($foreignKey, $localKey, $condition);
    }

    /**
     * 延迟获取关联数据
     * @param string   $subRelation 子关联名
     * @param \Closure $closure     闭包查询条件
     * @return false|\PDOStatement|string|\think\Collection
     */
    public function getRelation($subRelation = '', $closure = null)
    {
        if ($closure) {
            call_user_func_array($closure, [ & $this->query]);
        }
        $result = $this->buildQuery()->relation($subRelation)->select();
        $this->hydratePivot($result);
        return $result;
    }

    /**
     * 重载select方法
     * @param null $data
     * @return false|\PDOStatement|string|Collection
     */
    public function select($data = null)
    {
        $result = $this->buildQuery()->select($data);
        $this->hydratePivot($result);
        return $result;
    }

    /**
     * 重载paginate方法
     * @param null  $listRows
     * @param bool  $simple
     * @param array $config
     * @return Paginator
     */
    public function paginate($listRows = null, $simple = false, $config = [])
    {
        $result = $this->buildQuery()->paginate($listRows, $simple, $config);
        $this->hydratePivot($result);
        return $result;
    }

    /**
     * 重载find方法
     * @param null $data
     * @return array|false|\PDOStatement|string|Model
     */
    public function find($data = null)
    {
        $result = $this->buildQuery()->find($data);
        if ($result) {
            $this->hydratePivot([$result]);
        }
        return $result;
    }

    /**
     * 查找多条记录 如果不存在则抛出异常
     * @access public
     * @param array|string|Query|\Closure $data
     * @return array|\PDOStatement|string|Model
     */
    public function selectOrFail($data = null)
    {
        return $this->failException(true)->select($data);
    }

    /**
     * 查找单条记录 如果不存在则抛出异常
     * @access public
     * @param array|string|Query|\Closure $data
     * @return array|\PDOStatement|string|Model
     */
    public function findOrFail($data = null)
    {
        return $this->failException(true)->find($data);
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
     * @throws Exception
     */
    public function hasWhere($where = [])
    {
        throw new Exception('relation not support: hasWhere');
    }

    /**
     * 设置中间表的查询条件
     * @param      $field
     * @param null $op
     * @param null $condition
     * @return $this
     */
    public function wherePivot($field, $op = null, $condition = null)
    {
        $field = 'pivot.' . $field;
        $this->query->where($field, $op, $condition);
        return $this;
    }

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
        $localKey   = $this->localKey;
        $foreignKey = $this->foreignKey;

        $pk    = $resultSet[0]->getPk();
        $range = [];
        foreach ($resultSet as $result) {
            // 获取关联外键列表
            if (isset($result->$pk)) {
                $range[] = $result->$pk;
            }
        }

        if (!empty($range)) {
            // 查询关联数据
            $data = $this->eagerlyManyToMany([
                'pivot.' . $localKey => [
                    'in',
                    $range,
                ],
            ], $relation, $subRelation);
            // 关联属性名
            $attr = Loader::parseName($relation);
            // 关联数据封装
            foreach ($resultSet as $result) {
                if (!isset($data[$result->$pk])) {
                    $data[$result->$pk] = [];
                }

                $result->setRelation($attr, $this->resultSetBuild($data[$result->$pk]));
            }
        }
    }

    /**
     * 预载入关联查询（单个数据）
     * @access public
     * @param Model    $result      数据对象
     * @param string   $relation    当前关联名
     * @param string   $subRelation 子关联名
     * @param \Closure $closure     闭包
     * @return void
     */
    public function eagerlyResult(&$result, $relation, $subRelation, $closure)
    {
        $pk = $result->getPk();
        if (isset($result->$pk)) {
            $pk = $result->$pk;
            // 查询管理数据
            $data = $this->eagerlyManyToMany(['pivot.' . $this->localKey => $pk], $relation, $subRelation);

            // 关联数据封装
            if (!isset($data[$pk])) {
                $data[$pk] = [];
            }
            $result->setRelation(Loader::parseName($relation), $this->resultSetBuild($data[$pk]));
        }
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
        $pk    = $result->getPk();
        $count = 0;
        if (isset($result->$pk)) {
            $pk    = $result->$pk;
            $count = $this->belongsToManyQuery($this->foreignKey, $this->localKey, ['pivot.' . $this->localKey => $pk])->count();
        }
        return $count;
    }

    /**
     * 获取关联统计子查询
     * @access public
     * @param \Closure $closure 闭包
     * @return string
     */
    public function getRelationCountQuery($closure)
    {
        return $this->belongsToManyQuery($this->foreignKey, $this->localKey, [
            'pivot.' . $this->localKey => [
                'exp',
                '=' . $this->parent->getTable() . '.' . $this->parent->getPk(),
            ],
        ])->fetchSql()->count();
    }

    /**
     * 多对多 关联模型预查询
     * @access public
     * @param array  $where       关联预查询条件
     * @param string $relation    关联名
     * @param string $subRelation 子关联
     * @return array
     */
    protected function eagerlyManyToMany($where, $relation, $subRelation = '')
    {
        // 预载入关联查询 支持嵌套预载入
        $list = $this->belongsToManyQuery($this->foreignKey, $this->localKey, $where)->with($subRelation)->select();

        // 组装模型数据
        $data = [];
        foreach ($list as $set) {
            $pivot = [];
            foreach ($set->getData() as $key => $val) {
                if (strpos($key, '__')) {
                    list($name, $attr) = explode('__', $key, 2);
                    if ('pivot' == $name) {
                        $pivot[$attr] = $val;
                        unset($set->$key);
                    }
                }
            }
            $set->setRelation('pivot', $this->newPivot($pivot));
            $data[$pivot[$this->localKey]][] = $set;
        }
        return $data;
    }

    /**
     * BELONGS TO MANY 关联查询
     * @access public
     * @param string $foreignKey 关联模型关联键
     * @param string $localKey   当前模型关联键
     * @param array  $condition  关联查询条件
     * @return Query
     */
    protected function belongsToManyQuery($foreignKey, $localKey, $condition = [])
    {
        // 关联查询封装
        $tableName = $this->query->getTable();
        $table     = $this->pivot->getTable();
        $fields    = $this->getQueryFields($tableName);

        $query = $this->query->field($fields)
            ->field(true, false, $table, 'pivot', 'pivot__');

        if (empty($this->baseQuery)) {
            $relationFk = $this->query->getPk();
            $query->join($table . ' pivot', 'pivot.' . $foreignKey . '=' . $tableName . '.' . $relationFk)
                ->where($condition);
        }
        return $query;
    }

    /**
     * 保存（新增）当前关联数据对象
     * @access public
     * @param mixed $data  数据 可以使用数组 关联模型对象 和 关联对象的主键
     * @param array $pivot 中间表额外数据
     * @return integer
     */
    public function save($data, array $pivot = [])
    {
        // 保存关联表/中间表数据
        return $this->attach($data, $pivot);
    }

    /**
     * 批量保存当前关联数据对象
     * @access public
     * @param array $dataSet   数据集
     * @param array $pivot     中间表额外数据
     * @param bool  $samePivot 额外数据是否相同
     * @return integer
     */
    public function saveAll(array $dataSet, array $pivot = [], $samePivot = false)
    {
        $result = false;
        foreach ($dataSet as $key => $data) {
            if (!$samePivot) {
                $pivotData = isset($pivot[$key]) ? $pivot[$key] : [];
            } else {
                $pivotData = $pivot;
            }
            $result = $this->attach($data, $pivotData);
        }
        return $result;
    }

    /**
     * 附加关联的一个中间表数据
     * @access public
     * @param mixed $data  数据 可以使用数组、关联模型对象 或者 关联对象的主键
     * @param array $pivot 中间表额外数据
     * @return array|Pivot
     * @throws Exception
     */
    public function attach($data, $pivot = [])
    {
        if (is_array($data)) {
            if (key($data) === 0) {
                $id = $data;
            } else {
                // 保存关联表数据
                $model = new $this->model;
                $model->save($data);
                $id = $model->getLastInsID();
            }
        } elseif (is_numeric($data) || is_string($data)) {
            // 根据关联表主键直接写入中间表
            $id = $data;
        } elseif ($data instanceof Model) {
            // 根据关联表主键直接写入中间表
            $relationFk = $data->getPk();
            $id         = $data->$relationFk;
        }

        if ($id) {
            // 保存中间表数据
            $pk                     = $this->parent->getPk();
            $pivot[$this->localKey] = $this->parent->$pk;
            $ids                    = (array) $id;
            foreach ($ids as $id) {
                $pivot[$this->foreignKey] = $id;
                $this->pivot->insert($pivot, true);
                $result[] = $this->newPivot($pivot);
            }
            if (count($result) == 1) {
                // 返回中间表模型对象
                $result = $result[0];
            }
            return $result;
        } else {
            throw new Exception('miss relation data');
        }
    }

    /**
     * 解除关联的一个中间表数据
     * @access public
     * @param integer|array $data        数据 可以使用关联对象的主键
     * @param bool          $relationDel 是否同时删除关联表数据
     * @return integer
     */
    public function detach($data = null, $relationDel = false)
    {
        if (is_array($data)) {
            $id = $data;
        } elseif (is_numeric($data) || is_string($data)) {
            // 根据关联表主键直接写入中间表
            $id = $data;
        } elseif ($data instanceof Model) {
            // 根据关联表主键直接写入中间表
            $relationFk = $data->getPk();
            $id         = $data->$relationFk;
        }
        // 删除中间表数据
        $pk                     = $this->parent->getPk();
        $pivot[$this->localKey] = $this->parent->$pk;
        if (isset($id)) {
            $pivot[$this->foreignKey] = is_array($id) ? ['in', $id] : $id;
        }
        $this->pivot->where($pivot)->delete();
        // 删除关联表数据
        if (isset($id) && $relationDel) {
            $model = $this->model;
            $model::destroy($id);
        }
    }

    /**
     * 数据同步
     * @param array $ids
     * @param bool  $detaching
     * @return array
     */
    public function sync($ids, $detaching = true)
    {
        $changes = [
            'attached' => [],
            'detached' => [],
            'updated'  => [],
        ];
        $pk      = $this->parent->getPk();
        $current = $this->pivot->where($this->localKey, $this->parent->$pk)
            ->column($this->foreignKey);
        $records = [];

        foreach ($ids as $key => $value) {
            if (!is_array($value)) {
                $records[$value] = [];
            } else {
                $records[$key] = $value;
            }
        }

        $detach = array_diff($current, array_keys($records));

        if ($detaching && count($detach) > 0) {
            $this->detach($detach);

            $changes['detached'] = $detach;
        }

        foreach ($records as $id => $attributes) {
            if (!in_array($id, $current)) {
                $this->attach($id, $attributes);
                $changes['attached'][] = $id;
            } elseif (count($attributes) > 0 &&
                $this->attach($id, $attributes)
            ) {
                $changes['updated'][] = $id;
            }
        }

        return $changes;

    }

    /**
     * 执行基础查询（进执行一次）
     * @access protected
     * @return void
     */
    protected function baseQuery()
    {
        if (empty($this->baseQuery) && $this->parent->getData()) {
            $pk    = $this->parent->getPk();
            $table = $this->pivot->getTable();
            $this->query->join($table . ' pivot', 'pivot.' . $this->foreignKey . '=' . $this->query->getTable() . '.' . $this->query->getPk())->where('pivot.' . $this->localKey, $this->parent->$pk);
            $this->baseQuery = true;
        }
    }

}
