#!/bin/bash

if [ $(phpenv version-name) != "hhvm" ]; then
    cp tests/extensions/$(phpenv version-name)/*.so $(php-config --extension-dir)

    if [ $(phpenv version-name) = "7.0" ]; then
        phpenv config-add tests/conf/apcu_bc.ini
    else
        phpenv config-add tests/conf/apcu.ini
    fi

    phpenv config-add tests/conf/memcached.ini
    phpenv config-add tests/conf/redis.ini

    phpenv config-add tests/conf/timezone.ini
fi

composer install --no-interaction --ignore-platform-reqs
