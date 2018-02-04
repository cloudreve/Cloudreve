<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK IT ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006-2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: yunwuxin <448901948@qq.com>
// +----------------------------------------------------------------------

namespace think\console\output;

use think\console\Input;
use think\console\Output;
use think\console\output\question\Choice;
use think\console\output\question\Confirmation;

class Ask
{
    private static $stty;

    private static $shell;

    /** @var  Input */
    protected $input;

    /** @var  Output */
    protected $output;

    /** @var  Question */
    protected $question;

    public function __construct(Input $input, Output $output, Question $question)
    {
        $this->input    = $input;
        $this->output   = $output;
        $this->question = $question;
    }

    public function run()
    {
        if (!$this->input->isInteractive()) {
            return $this->question->getDefault();
        }

        if (!$this->question->getValidator()) {
            return $this->doAsk();
        }

        $that = $this;

        $interviewer = function () use ($that) {
            return $that->doAsk();
        };

        return $this->validateAttempts($interviewer);
    }

    protected function doAsk()
    {
        $this->writePrompt();

        $inputStream  = STDIN;
        $autocomplete = $this->question->getAutocompleterValues();

        if (null === $autocomplete || !$this->hasSttyAvailable()) {
            $ret = false;
            if ($this->question->isHidden()) {
                try {
                    $ret = trim($this->getHiddenResponse($inputStream));
                } catch (\RuntimeException $e) {
                    if (!$this->question->isHiddenFallback()) {
                        throw $e;
                    }
                }
            }

            if (false === $ret) {
                $ret = fgets($inputStream, 4096);
                if (false === $ret) {
                    throw new \RuntimeException('Aborted');
                }
                $ret = trim($ret);
            }
        } else {
            $ret = trim($this->autocomplete($inputStream));
        }

        $ret = strlen($ret) > 0 ? $ret : $this->question->getDefault();

        if ($normalizer = $this->question->getNormalizer()) {
            return $normalizer($ret);
        }

        return $ret;
    }

    private function autocomplete($inputStream)
    {
        $autocomplete = $this->question->getAutocompleterValues();
        $ret          = '';

        $i          = 0;
        $ofs        = -1;
        $matches    = $autocomplete;
        $numMatches = count($matches);

        $sttyMode = shell_exec('stty -g');

        shell_exec('stty -icanon -echo');

        while (!feof($inputStream)) {
            $c = fread($inputStream, 1);

            if ("\177" === $c) {
                if (0 === $numMatches && 0 !== $i) {
                    --$i;
                    $this->output->write("\033[1D");
                }

                if ($i === 0) {
                    $ofs        = -1;
                    $matches    = $autocomplete;
                    $numMatches = count($matches);
                } else {
                    $numMatches = 0;
                }

                $ret = substr($ret, 0, $i);
            } elseif ("\033" === $c) {
                $c .= fread($inputStream, 2);

                if (isset($c[2]) && ('A' === $c[2] || 'B' === $c[2])) {
                    if ('A' === $c[2] && -1 === $ofs) {
                        $ofs = 0;
                    }

                    if (0 === $numMatches) {
                        continue;
                    }

                    $ofs += ('A' === $c[2]) ? -1 : 1;
                    $ofs = ($numMatches + $ofs) % $numMatches;
                }
            } elseif (ord($c) < 32) {
                if ("\t" === $c || "\n" === $c) {
                    if ($numMatches > 0 && -1 !== $ofs) {
                        $ret = $matches[$ofs];
                        $this->output->write(substr($ret, $i));
                        $i = strlen($ret);
                    }

                    if ("\n" === $c) {
                        $this->output->write($c);
                        break;
                    }

                    $numMatches = 0;
                }

                continue;
            } else {
                $this->output->write($c);
                $ret .= $c;
                ++$i;

                $numMatches = 0;
                $ofs        = 0;

                foreach ($autocomplete as $value) {
                    if (0 === strpos($value, $ret) && $i !== strlen($value)) {
                        $matches[$numMatches++] = $value;
                    }
                }
            }

            $this->output->write("\033[K");

            if ($numMatches > 0 && -1 !== $ofs) {
                $this->output->write("\0337");
                $this->output->highlight(substr($matches[$ofs], $i));
                $this->output->write("\0338");
            }
        }

        shell_exec(sprintf('stty %s', $sttyMode));

        return $ret;
    }

    protected function getHiddenResponse($inputStream)
    {
        if ('\\' === DIRECTORY_SEPARATOR) {
            $exe = __DIR__ . '/../bin/hiddeninput.exe';

            $value = rtrim(shell_exec($exe));
            $this->output->writeln('');

            if (isset($tmpExe)) {
                unlink($tmpExe);
            }

            return $value;
        }

        if ($this->hasSttyAvailable()) {
            $sttyMode = shell_exec('stty -g');

            shell_exec('stty -echo');
            $value = fgets($inputStream, 4096);
            shell_exec(sprintf('stty %s', $sttyMode));

            if (false === $value) {
                throw new \RuntimeException('Aborted');
            }

            $value = trim($value);
            $this->output->writeln('');

            return $value;
        }

        if (false !== $shell = $this->getShell()) {
            $readCmd = $shell === 'csh' ? 'set mypassword = $<' : 'read -r mypassword';
            $command = sprintf("/usr/bin/env %s -c 'stty -echo; %s; stty echo; echo \$mypassword'", $shell, $readCmd);
            $value   = rtrim(shell_exec($command));
            $this->output->writeln('');

            return $value;
        }

        throw new \RuntimeException('Unable to hide the response.');
    }

    protected function validateAttempts($interviewer)
    {
        /** @var \Exception $error */
        $error    = null;
        $attempts = $this->question->getMaxAttempts();
        while (null === $attempts || $attempts--) {
            if (null !== $error) {
                $this->output->error($error->getMessage());
            }

            try {
                return call_user_func($this->question->getValidator(), $interviewer());
            } catch (\Exception $error) {
            }
        }

        throw $error;
    }

    /**
     * 显示问题的提示信息
     */
    protected function writePrompt()
    {
        $text    = $this->question->getQuestion();
        $default = $this->question->getDefault();

        switch (true) {
            case null === $default:
                $text = sprintf(' <info>%s</info>:', $text);

                break;

            case $this->question instanceof Confirmation:
                $text = sprintf(' <info>%s (yes/no)</info> [<comment>%s</comment>]:', $text, $default ? 'yes' : 'no');

                break;

            case $this->question instanceof Choice && $this->question->isMultiselect():
                $choices = $this->question->getChoices();
                $default = explode(',', $default);

                foreach ($default as $key => $value) {
                    $default[$key] = $choices[trim($value)];
                }

                $text = sprintf(' <info>%s</info> [<comment>%s</comment>]:', $text, implode(', ', $default));

                break;

            case $this->question instanceof Choice:
                $choices = $this->question->getChoices();
                $text    = sprintf(' <info>%s</info> [<comment>%s</comment>]:', $text, $choices[$default]);

                break;

            default:
                $text = sprintf(' <info>%s</info> [<comment>%s</comment>]:', $text, $default);
        }

        $this->output->writeln($text);

        if ($this->question instanceof Choice) {
            $width = max(array_map('strlen', array_keys($this->question->getChoices())));

            foreach ($this->question->getChoices() as $key => $value) {
                $this->output->writeln(sprintf("  [<comment>%-${width}s</comment>] %s", $key, $value));
            }
        }

        $this->output->write(' > ');
    }

    private function getShell()
    {
        if (null !== self::$shell) {
            return self::$shell;
        }

        self::$shell = false;

        if (file_exists('/usr/bin/env')) {
            $test = "/usr/bin/env %s -c 'echo OK' 2> /dev/null";
            foreach (['bash', 'zsh', 'ksh', 'csh'] as $sh) {
                if ('OK' === rtrim(shell_exec(sprintf($test, $sh)))) {
                    self::$shell = $sh;
                    break;
                }
            }
        }

        return self::$shell;
    }

    private function hasSttyAvailable()
    {
        if (null !== self::$stty) {
            return self::$stty;
        }

        exec('stty 2>&1', $output, $exitcode);

        return self::$stty = $exitcode === 0;
    }
}
