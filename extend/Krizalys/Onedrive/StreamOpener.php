<?php

namespace Krizalys\Onedrive;

class StreamOpener
{
    private static $uris = [
        StreamBackEnd::MEMORY => 'php://memory',
        StreamBackEnd::TEMP   => 'php://temp',
    ];

    /**
     * Opens a stream given a stream back end.
     *
     * @param int $streamBackEnd The stream back end.
     *
     * @return bool|resource The open stream.
     *
     * @throws \Exception Thrown if the stream back end given is not supported.
     */
    public function open($streamBackEnd)
    {
        if (!array_key_exists($streamBackEnd, self::$uris)) {
            throw new \Exception("Unsupported stream back end: $streamBackEnd");
        }

        $uri = self::$uris[$streamBackEnd];
        return fopen($uri, 'rw+b');
    }
}
