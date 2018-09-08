<?php

namespace Krizalys\Onedrive;

class StreamBackEnd
{
    // Memory-backed stream.
    const MEMORY = 1;

    // Temporary file-backed stream. A temporary file is actually used if the
    // stream contents exceeds 2 MiB.
    const TEMP = 2;
}
