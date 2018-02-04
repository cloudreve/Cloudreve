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

namespace think\console\output\question;

use think\console\output\Question;

class Choice extends Question
{

    private $choices;
    private $multiselect  = false;
    private $prompt       = ' > ';
    private $errorMessage = 'Value "%s" is invalid';

    /**
     * 构造方法
     * @param string $question 问题
     * @param array  $choices  选项
     * @param mixed  $default  默认答案
     */
    public function __construct($question, array $choices, $default = null)
    {
        parent::__construct($question, $default);

        $this->choices = $choices;
        $this->setValidator($this->getDefaultValidator());
        $this->setAutocompleterValues($choices);
    }

    /**
     * 可选项
     * @return array
     */
    public function getChoices()
    {
        return $this->choices;
    }

    /**
     * 设置可否多选
     * @param bool $multiselect
     * @return self
     */
    public function setMultiselect($multiselect)
    {
        $this->multiselect = $multiselect;
        $this->setValidator($this->getDefaultValidator());

        return $this;
    }

    public function isMultiselect()
    {
        return $this->multiselect;
    }

    /**
     * 获取提示
     * @return string
     */
    public function getPrompt()
    {
        return $this->prompt;
    }

    /**
     * 设置提示
     * @param string $prompt
     * @return self
     */
    public function setPrompt($prompt)
    {
        $this->prompt = $prompt;

        return $this;
    }

    /**
     * 设置错误提示信息
     * @param string $errorMessage
     * @return self
     */
    public function setErrorMessage($errorMessage)
    {
        $this->errorMessage = $errorMessage;
        $this->setValidator($this->getDefaultValidator());

        return $this;
    }

    /**
     * 获取默认的验证方法
     * @return callable
     */
    private function getDefaultValidator()
    {
        $choices      = $this->choices;
        $errorMessage = $this->errorMessage;
        $multiselect  = $this->multiselect;
        $isAssoc      = $this->isAssoc($choices);

        return function ($selected) use ($choices, $errorMessage, $multiselect, $isAssoc) {
            // Collapse all spaces.
            $selectedChoices = str_replace(' ', '', $selected);

            if ($multiselect) {
                // Check for a separated comma values
                if (!preg_match('/^[a-zA-Z0-9_-]+(?:,[a-zA-Z0-9_-]+)*$/', $selectedChoices, $matches)) {
                    throw new \InvalidArgumentException(sprintf($errorMessage, $selected));
                }
                $selectedChoices = explode(',', $selectedChoices);
            } else {
                $selectedChoices = [$selected];
            }

            $multiselectChoices = [];
            foreach ($selectedChoices as $value) {
                $results = [];
                foreach ($choices as $key => $choice) {
                    if ($choice === $value) {
                        $results[] = $key;
                    }
                }

                if (count($results) > 1) {
                    throw new \InvalidArgumentException(sprintf('The provided answer is ambiguous. Value should be one of %s.', implode(' or ', $results)));
                }

                $result = array_search($value, $choices);

                if (!$isAssoc) {
                    if (!empty($result)) {
                        $result = $choices[$result];
                    } elseif (isset($choices[$value])) {
                        $result = $choices[$value];
                    }
                } elseif (empty($result) && array_key_exists($value, $choices)) {
                    $result = $value;
                }

                if (empty($result)) {
                    throw new \InvalidArgumentException(sprintf($errorMessage, $value));
                }
                array_push($multiselectChoices, $result);
            }

            if ($multiselect) {
                return $multiselectChoices;
            }

            return current($multiselectChoices);
        };
    }
}
