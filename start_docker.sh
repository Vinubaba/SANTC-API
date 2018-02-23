#!/bin/bash

docker run --name POSTGRES -v /home/agustin/gocode/src/github.com/DigitalFrameworksLLC/teddycare/sql/:/home/agustin/gocode/src/github.com/DigitalFrameworksLLC/teddycare/sql/ -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=teddycare postgres
