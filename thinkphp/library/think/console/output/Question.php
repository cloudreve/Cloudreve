<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2015 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: yunwuxin <448901948@qq.com>
// +----------------------------------------------------------------------

namespace think\console\output;

class Question
{

    private $question;
    private $attempts;
    private $hidden         = false;
    private $hiddenFallback = true;
    private $autocompleterValues;
    private $validator;
    private $default;
    private $normalizer;

    /**
     * 构造方法
     * @param string $question 问题
     * @param mixed  $default  默认答案
     */
    public function __construct($question, $default = null)
    {
        $this->question = $question;
        $this->default  = $default;
    }

    /**
     * 获取问题
     * @return string
     */
    public function getQuestion()
    {
        return $this->question;
    }

    /**
     * 获取默认答案
     * @return mixed
     */
    public function getDefault()
    {
        return $this->default;
    }

    /**
     * 是否隐藏答案
     * @return bool
     */
    public function isHidden()
    {
        return $this->hidden;
    }

    /**
     * 隐藏答案
     * @param bool $hidden
     * @return Question
     */
    public function setHidden($hidden)
    {
        if ($this->autocompleterValues) {
            throw new \LogicException('A hidden question cannot use the autocompleter.');
        }

        $this->hidden = (bool) $hidden;

        return $this;
    }

    /**
     * 不能被隐藏是否撤销
     * @return bool
     */
    public function isHiddenFallback()
    {
        return $this->hiddenFallback;
    }

    /**
     * 设置不能被隐藏的时候的操作
     * @param bool $fallback
     * @return Question
     */
    public function setHiddenFallback($fallback)
    {
        $this->hiddenFallback = (bool) $fallback;

        return $this;
    }

    /**
     * 获取自动完成
     * @return null|array|\Traversable
     */
    public function getAutocompleterValues()
    {
        return $this->autocompleterValues;
    }

    /**
     * 设置自动完成的值
     * @param null|array|\Traversable $values
     * @return Question
     * @throws \InvalidArgumentException
     * @throws \LogicException
     */
    public function setAutocompleterValues($values)
    {
        if (is_array($values) && $this->isAssoc($values)) {
            $values = array_merge(array_keys($values), array_values($values));
        }

        if (null !== $values && !is_array($values)) {
            if (!$values instanceof \Traversable || $values instanceof \Countable) {
                throw new \InvalidArgumentException('Autocompleter values can be either an array, `null` or an object implementing both `Countable` and `Traversable` interfaces.');
            }
        }

        if ($this->hidden) {
            throw new \LogicException('A hidden question cannot use the autocompleter.');
        }

        $this->autocompleterValues = $values;

        return $this;
    }

    /**
     * 设置答案的验证器
     * @param null|callable $validator
     * @return Question The current instance
     */
    public function setValidator($validator)
    {
        $this->validator = $validator;

        return $this;
    }

    /**
     * 获取验证器
     * @return null|callable
     */
    public function getValidator()
    {
        return $this->validator;
    }

    /**
     * 设置最大重试次数
     * @param null|int $attempts
     * @return Question
     * @throws \InvalidArgumentException
     */
    public function setMaxAttempts($attempts)
    {
        if (null !== $attempts && $attempts < 1) {
            throw new \InvalidArgumentException('Maximum number of attempts must be a positive value.');
        }

        $this->attempts = $attempts;

        return $this;
    }

    /**
     * 获取最大重试次数
     * @return null|int
     */
    public function getMaxAttempts()
    {
        return $this->attempts;
    }

    /**
     * 设置响应的回调
     * @param string|\Closure $normalizer
     * @return Question
     */
    public function setNormalizer($normalizer)
    {
        $this->normalizer = $normalizer;

        return $this;
    }

    /**
     * 获取响应回调
     * The normalizer can ba a callable (a string), a closure or a class implementing __invoke.
     * @return string|\Closure
     */
    public function getNormalizer()
    {
        return $this->normalizer;
    }

    protected function isAssoc($array)
    {
        return (bool) count(array_filter(array_keys($array), 'is_string'));
    }
}
