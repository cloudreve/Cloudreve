<?php

namespace traits\model;

use think\db\Query;

trait SoftDelete
{

    /**
     * 判断当前实例是否被软删除
     * @access public
     * @return boolean
     */
    public function trashed()
    {
        $field = $this->getDeleteTimeField();
        if (!empty($this->data[$field])) {
            return true;
        }
        return false;
    }

    /**
     * 查询软删除数据
     * @access public
     * @return Query
     */
    public static function withTrashed()
    {
        $model = new static();
        $field = $model->getDeleteTimeField(true);
        return $model->getQuery();
    }

    /**
     * 只查询软删除数据
     * @access public
     * @return Query
     */
    public static function onlyTrashed()
    {
        $model = new static();
        $field = $model->getDeleteTimeField(true);
        return $model->getQuery()
            ->useSoftDelete($field, ['not null', '']);
    }

    /**
     * 删除当前的记录
     * @access public
     * @param bool  $force 是否强制删除
     * @return integer
     */
    public function delete($force = false)
    {
        if (false === $this->trigger('before_delete', $this)) {
            return false;
        }
        $name = $this->getDeleteTimeField();
        if (!$force) {
            // 软删除
            $this->data[$name] = $this->autoWriteTimestamp($name);
            $result            = $this->isUpdate()->save();
        } else {
            $result = $this->getQuery()->delete($this->data);
        }

        $this->trigger('after_delete', $this);
        return $result;
    }

    /**
     * 删除记录
     * @access public
     * @param mixed $data 主键列表 支持闭包查询条件
     * @param bool  $force 是否强制删除
     * @return integer 成功删除的记录数
     */
    public static function destroy($data, $force = false)
    {
        // 包含软删除数据
        $query = self::withTrashed();
        if (is_array($data) && key($data) !== 0) {
            $query->where($data);
            $data = null;
        } elseif ($data instanceof \Closure) {
            call_user_func_array($data, [ & $query]);
            $data = null;
        } elseif (is_null($data)) {
            return 0;
        }

        $resultSet = $query->select($data);
        $count     = 0;
        if ($resultSet) {
            foreach ($resultSet as $data) {
                $result = $data->delete($force);
                $count += $result;
            }
        }
        return $count;
    }

    /**
     * 恢复被软删除的记录
     * @access public
     * @param array $where 更新条件
     * @return integer
     */
    public function restore($where = [])
    {
        $name = $this->getDeleteTimeField();
        if (empty($where)) {
            $pk         = $this->getPk();
            $where[$pk] = $this->getData($pk);
        }
        // 恢复删除
        return $this->getQuery()
            ->useSoftDelete($name, ['not null', ''])
            ->where($where)
            ->update([$name => null]);
    }

    /**
     * 查询默认不包含软删除数据
     * @access protected
     * @param Query $query 查询对象
     * @return void
     */
    protected function base($query)
    {
        $field = $this->getDeleteTimeField(true);
        $query->useSoftDelete($field);
    }

    /**
     * 获取软删除字段
     * @access public
     * @param bool  $read 是否查询操作 写操作的时候会自动去掉表别名
     * @return string
     */
    protected function getDeleteTimeField($read = false)
    {
        $field = property_exists($this, 'deleteTime') && isset($this->deleteTime) ? $this->deleteTime : 'delete_time';
        if (!strpos($field, '.')) {
            $field = '__TABLE__.' . $field;
        }
        if (!$read && strpos($field, '.')) {
            $array = explode('.', $field);
            $field = array_pop($array);
        }
        return $field;
    }
}
