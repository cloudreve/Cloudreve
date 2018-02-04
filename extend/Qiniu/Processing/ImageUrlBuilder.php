<?php
namespace Qiniu\Processing;

use Qiniu;

/**
 * 主要涉及图片链接拼接
 *
 * @link http://developer.qiniu.com/code/v6/api/kodo-api/image/imageview2.html
 */
final class ImageUrlBuilder
{
    /**
     * mode合法范围值
     *
     * @var array
     */
    protected $modeArr = array(0, 1, 2, 3, 4, 5);

    /**
     * format合法值
     *
     * @var array
     */
    protected $formatArr = array('psd', 'jpeg', 'png', 'gif', 'webp', 'tiff', 'bmp');

    /**
     * 水印图片位置合法值
     *
     * @var array
     */
    protected $gravityArr = array('NorthWest', 'North', 'NorthEast',
        'West', 'Center', 'East', 'SouthWest', 'South', 'SouthEast');

    /**
     * 缩略图链接拼接
     *
     * @param  string $url 图片链接
     * @param  int $mode 缩略模式
     * @param  int $width 宽度
     * @param  int $height 长度
     * @param  string $format 输出类型
     * @param  int $quality 图片质量
     * @param  int $interlace 是否支持渐进显示
     * @param  int $ignoreError 忽略结果
     * @return string
     * @link http://developer.qiniu.com/code/v6/api/kodo-api/image/imageview2.html
     * @author Sherlock Ren <sherlock_ren@icloud.com>
     */
    public function thumbnail(
        $url,
        $mode,
        $width,
        $height,
        $format = null,
        $interlace = null,
        $quality = null,
        $ignoreError = 1
    ) {
    
        // url合法效验
        if (! $this->isUrl($url)) {
            return $url;
        }

        // 参数合法性效验
        if (! in_array(intval($mode), $this->modeArr, true)) {
            return $url;
        }

        if (! $width || ! $height) {
            return $url;
        }

        $thumbStr = 'imageView2/' . $mode . '/w/' . $width . '/h/' . $height . '/';

        // 拼接输出格式
        if (! is_null($format)
            && in_array($format, $this->formatArr)
        ) {
            $thumbStr .= 'format/' . $format . '/';
        }

        // 拼接渐进显示
        if (! is_null($interlace)
            && in_array(intval($interlace), array(0, 1), true)
        ) {
            $thumbStr .= 'interlace/' . $interlace . '/';
        }

        // 拼接图片质量
        if (! is_null($quality)
            && intval($quality) >= 0
            && intval($quality) <= 100
        ) {
            $thumbStr .= 'q/' . $quality . '/';
        }

        $thumbStr .= 'ignore-error/' . $ignoreError . '/';

        // 如果有query_string用|线分割实现多参数
        return $url . ($this->hasQuery($url) ? '|' : '?') . $thumbStr;
    }

    /**
     * 图片水印
     *
     * @param  string $url 图片链接
     * @param  string $image 水印图片链接
     * @param  numeric $dissolve 透明度
     * @param  string $gravity 水印位置
     * @param  numeric $dx 横轴边距
     * @param  numeric $dy 纵轴边距
     * @param  numeric $watermarkScale 自适应原图的短边比例
     * @link   http://developer.qiniu.com/code/v6/api/kodo-api/image/watermark.html
     * @return string
     * @author Sherlock Ren <sherlock_ren@icloud.com>
     */
    public function waterImg(
        $url,
        $image,
        $dissolve = 100,
        $gravity = 'SouthEast',
        $dx = null,
        $dy = null,
        $watermarkScale = null
    ) {
        // url合法效验
        if (! $this->isUrl($url)) {
            return $url;
        }

        $waterStr = 'watermark/1/image/' . \Qiniu\base64_urlSafeEncode($image) . '/';

        // 拼接水印透明度
        if (is_numeric($dissolve)
            && $dissolve <= 100
        ) {
            $waterStr .= 'dissolve/' . $dissolve . '/';
        }

        // 拼接水印位置
        if (in_array($gravity, $this->gravityArr, true)) {
            $waterStr .= 'gravity/' . $gravity . '/';
        }

        // 拼接横轴边距
        if (! is_null($dx)
            && is_numeric($dx)
        ) {
            $waterStr .= 'dx/' . $dx . '/';
        }

        // 拼接纵轴边距
        if (! is_null($dy)
            && is_numeric($dy)
        ) {
            $waterStr .= 'dy/' . $dy . '/';
        }

        // 拼接自适应原图的短边比例
        if (! is_null($watermarkScale)
            && is_numeric($watermarkScale)
            && $watermarkScale > 0
            && $watermarkScale < 1
        ) {
            $waterStr .= 'ws/' . $watermarkScale . '/';
        }

        // 如果有query_string用|线分割实现多参数
        return $url . ($this->hasQuery($url) ? '|' : '?') . $waterStr;
    }

    /**
     * 文字水印
     *
     * @param  string $url 图片链接
     * @param  string $text 文字
     * @param  string $font 文字字体
     * @param  string $fontSize 文字字号
     * @param  string $fontColor 文字颜色
     * @param  numeric $dissolve 透明度
     * @param  string $gravity 水印位置
     * @param  numeric $dx 横轴边距
     * @param  numeric $dy 纵轴边距
     * @link   http://developer.qiniu.com/code/v6/api/kodo-api/image/watermark.html#text-watermark
     * @return string
     * @author Sherlock Ren <sherlock_ren@icloud.com>
     */
    public function waterText(
        $url,
        $text,
        $font = '黑体',
        $fontSize = 0,
        $fontColor = null,
        $dissolve = 100,
        $gravity = 'SouthEast',
        $dx = null,
        $dy = null
    ) {
        // url合法效验
        if (! $this->isUrl($url)) {
            return $url;
        }

        $waterStr = 'watermark/2/text/'
            . \Qiniu\base64_urlSafeEncode($text) . '/font/'
            . \Qiniu\base64_urlSafeEncode($font) . '/';

        // 拼接文字大小
        if (is_int($fontSize)) {
            $waterStr .= 'fontsize/' . $fontSize . '/';
        }

        // 拼接文字颜色
        if (! is_null($fontColor)
            && $fontColor
        ) {
            $waterStr .= 'fill/' . \Qiniu\base64_urlSafeEncode($fontColor) . '/';
        }

        // 拼接水印透明度
        if (is_numeric($dissolve)
            && $dissolve <= 100
        ) {
            $waterStr .= 'dissolve/' . $dissolve . '/';
        }

        // 拼接水印位置
        if (in_array($gravity, $this->gravityArr, true)) {
            $waterStr .= 'gravity/' . $gravity . '/';
        }

        // 拼接横轴边距
        if (! is_null($dx)
            && is_numeric($dx)
        ) {
            $waterStr .= 'dx/' . $dx . '/';
        }

        // 拼接纵轴边距
        if (! is_null($dy)
            && is_numeric($dy)
        ) {
            $waterStr .= 'dy/' . $dy . '/';
        }

        // 如果有query_string用|线分割实现多参数
        return $url . ($this->hasQuery($url) ? '|' : '?') . $waterStr;
    }

    /**
     * 效验url合法性
     *
     * @param  string $url url链接
     * @return string
     * @author Sherlock Ren <sherlock_ren@icloud.com>
     */
    protected function isUrl($url)
    {
        $urlArr = parse_url($url);

        return $urlArr['scheme']
            && in_array($urlArr['scheme'], array('http', 'https'))
            && $urlArr['host']
            && $urlArr['path'];
    }

    /**
     * 检测是否有query
     *
     * @param  string $url url链接
     * @return string
     * @author Sherlock Ren <sherlock_ren@icloud.com>
     */
    protected function hasQuery($url)
    {
        $urlArr = parse_url($url);

        return ! empty($urlArr['query']);
    }
}
